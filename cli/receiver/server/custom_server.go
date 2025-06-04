package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage"
	"github.com/kuznetsovin/egts-protocol/libs/egts"
	log "github.com/sirupsen/logrus"
)

type CustomServer struct {
	Server
}

func NewCustom(addr string, ttl time.Duration, store storage.Saver) *CustomServer {
	return &CustomServer{Server: New(addr, ttl, store)}
}

func (server *CustomServer) Run() {
	var err error
	server.l, err = net.Listen("tcp", server.addr)
	if err != nil {
		log.Fatalf("Не удалось открыть соединение: %v", err)
	}
	defer server.l.Close()

	log.WithField("addr", server.addr).Info("Запущен сервер")
	log.Debug("TTL: ", server.ttl)

	for {
		conn, err := server.l.Accept()
		if err != nil {
			log.WithField("err", err).Error("Ошибка соединения")
			continue
		}

		go server.handleConn(conn)
	}
}

func (s *CustomServer) handleConn(conn net.Conn) {
	defer conn.Close()

	if s.store == nil {
		log.Error("Некорректная ссылка на объект хранилища")
		return
	}

	log.WithField("ip", conn.RemoteAddr()).Info("Установлено соединение")

	for {
		packet, err := s.readPacket(conn)
		if err != nil {
			return
		}

		pkg, receivedTimestamp, resultCode, err := s.decodePacket(packet)
		if err != nil {
			s.sendDecodeError(conn, pkg.PacketIdentifier, resultCode)
			continue
		}

		switch pkg.PacketType {
		case egts.PtAppdataPacket:
			if err := s.handleAppData(conn, pkg, receivedTimestamp, resultCode); err != nil {
				continue
			}
		case egts.PtResponsePacket:
			log.Debug("Тип пакета EGTS_PT_RESPONSE")
		}
	}
}

func (s *CustomServer) readPacket(conn net.Conn) ([]byte, error) {
	if s.ttl > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.ttl))
	} else {
		_ = conn.SetReadDeadline(time.Time{})
	}

	headerBuf := make([]byte, headerLen)
	_, err := io.ReadFull(conn, headerBuf)
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			log.WithField("ip", conn.RemoteAddr()).Warn("Таймаут чтения")
		} else if err == io.EOF {
			log.WithField("ip", conn.RemoteAddr()).Info("Клиент закрыл соединение")
		} else {
			log.WithField("err", err).Error("Ошибка при получении")
		}
		_ = conn.SetDeadline(time.Time{})
		return nil, err
	}

	if headerBuf[0] != 0x01 {
		log.WithField("ip", conn.RemoteAddr()).Warn("Пакет не соответствует формату EGTS")
		_ = conn.SetDeadline(time.Time{})
		return nil, fmt.Errorf("invalid EGTS packet")
	}

	bodyLen := binary.LittleEndian.Uint16(headerBuf[5:7])
	pkgLen := uint16(headerBuf[3])
	if bodyLen > 0 {
		pkgLen += bodyLen + 2
	}

	if s.ttl > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.ttl))
	} else {
		_ = conn.SetReadDeadline(time.Time{})
	}

	buf := make([]byte, pkgLen-headerLen)
	if _, err := io.ReadFull(conn, buf); err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			log.WithField("ip", conn.RemoteAddr()).Warn("Таймаут чтения")
		} else {
			log.WithField("err", err).Error("Ошибка при получении тела пакета")
		}
		_ = conn.SetDeadline(time.Time{})
		return nil, err
	}

	_ = conn.SetReadDeadline(time.Time{})
	packet := append(headerBuf, buf...)
	log.Debug("Принят пакет")
	return packet, nil
}

func (s *CustomServer) decodePacket(packet []byte) (*egts.Package, int64, uint8, error) {
	pkg := egts.Package{}
	receivedTimestamp := time.Now().UTC().Unix()
	resultCode, err := pkg.Decode(packet)
	return &pkg, receivedTimestamp, resultCode, err
}

func (s *CustomServer) sendDecodeError(conn net.Conn, packetIdentifier uint16, resultCode uint8) {
	resp, err := createPtResponse(packetIdentifier, resultCode, 0, nil)
	if err != nil {
		log.WithField("err", err).Error("Ошибка сборки ответа EGTS_PT_RESPONSE с ошибкой")
		return
	}
	_, _ = conn.Write(resp)
}

func (s *CustomServer) handleAppData(conn net.Conn, pkg *egts.Package, receivedTimestamp int64, resultCode uint8) error {
	var (
		srResponsesRecord egts.RecordDataSet
		srResultCodePkg   []byte
		serviceType       uint8
		client            uint32
	)

	for _, rec := range *pkg.ServicesFrameData.(*egts.ServiceDataSet) {
		exportPacket := storage.NavRecord{
			PacketID: uint32(pkg.PacketIdentifier),
		}

		// TODO: узнать, нужно ли проверять Recipient Service Type
		serviceType = rec.SourceServiceType
		log.Debug("Тип сервиса: ", serviceType)

		if serviceType != 2 {
			log.Warn("Неподдерживаемый сервис")
			srResponsesRecord = append(srResponsesRecord, egts.RecordData{
				SubrecordType:   egts.SrRecordResponseType,
				SubrecordLength: 3,
				SubrecordData: &egts.SrResponse{
					ConfirmedRecordNumber: rec.RecordNumber,
					RecordStatus:          egtsPcSrvcDenied,
				},
			})

			continue
		}

		if rec.ObjectIDFieldExists == "1" {
			client = rec.ObjectIdentifier
		}

		var (
			recStatus uint8 = egtsPcOk
			isPkgSave bool  = false
		)

		for _, subRec := range rec.RecordDataSet {
			switch subRecData := subRec.SubrecordData.(type) {
			case *egts.SrResponse:
				log.Debug("Встречена подзапись EGTS_SR_RESPONSE")
			case *egts.SrAdSensorsData:
				log.Debug("Встречена подзапись EGTS_SR_AD_SENSORS_DATA")
			case *egts.SrCountersData:
				log.Debug("Встречена подзапись EGTS_SR_COUNTERS_DATA")
			case *egts.SrStateData:
				log.Debug("Встречена подзапись EGTS_SR_STATE_DATA")
			case *egts.SrAbsAnSensData:
				log.Debug("Встречена подзапись EGTS_SR_ABS_AN_SENS_DATA")
			case *egts.SrAbsCntrData:
				log.Debug("Встречена подзапись EGTS_SR_ABS_CNTR_DATA")
			case *egts.SrLiquidLevelSensor:
				log.Debug("Встречена подзапись EGTS_SR_LIQUID_LEVEL_SENSOR")
			case *egts.SrPassengersCountersData:
				log.Debug("Встречена подзапись EGTS_SR_PASSENGERS_COUNTERS_DATA")
			case *egts.SrLoopinData:
				log.Debug("Встречена подзапись EGTS_SR_LOOPIN_DATA")
			case *egts.SrAbsDigSensData:
				log.Debug("Встречена подзапись EGTS_SR_ABS_DIG_SENS_DATA")
			case *egts.SrAbsLoopinData:
				log.Debug("Встречена подзапись EGTS_SR_ABS_LOOPIN_DATA")
			case *egts.SrPosData:
				log.Debug("Разбор подзаписи EGTS_SR_POS_DATA")
				isPkgSave = true
				exportPacket.NavigationTimestamp = subRecData.NavigationTime.Unix()
				exportPacket.ReceivedTimestamp = receivedTimestamp
				exportPacket.Latitude = subRecData.Latitude
				exportPacket.Longitude = subRecData.Longitude
				exportPacket.Speed = subRecData.Speed
				exportPacket.Course = subRecData.Direction
			case *egts.SrExtPosData:
				log.Debug("Разбор подзаписи EGTS_SR_EXT_POS_DATA")
				exportPacket.Nsat = subRecData.Satellites
				exportPacket.Pdop = subRecData.PositionDilutionOfPrecision
				exportPacket.Hdop = subRecData.HorizontalDilutionOfPrecision
				exportPacket.Vdop = subRecData.VerticalDilutionOfPrecision
				exportPacket.Ns = subRecData.NavigationSystem
			default:
				log.Warnf("Неподдерживаемая подзапись SRT=%d в записи RN=%d",
					subRec.SubrecordType, rec.RecordNumber)
				recStatus = egtsPcUnsType
			}
		}

		srResponsesRecord = append(srResponsesRecord, egts.RecordData{
			SubrecordType:   egts.SrRecordResponseType,
			SubrecordLength: 3,
			SubrecordData: &egts.SrResponse{
				ConfirmedRecordNumber: rec.RecordNumber,
				RecordStatus:          recStatus,
			},
		})

		exportPacket.Client = client
		if isPkgSave && recStatus == egtsPcOk {
			if err := s.store.Save(&exportPacket); err != nil {
				log.WithField("err", err).Error("Ошибка сохранения телеметрии")
			}
		}
	}

	resp, err := createPtResponse(pkg.PacketIdentifier, resultCode, serviceType, srResponsesRecord)
	if err != nil {
		log.WithField("err", err).Error("Ошибка сборки ответа")
		return err
	}
	_, _ = conn.Write(resp)
	log.Debug("Отправлен пакет EGTS_PT_RESPONSE")

	if len(srResultCodePkg) > 0 {
		_, _ = conn.Write(srResultCodePkg)
		log.Debug("Отправлен пакет EGTS_SR_RESULT_CODE")
	}

	return nil
}
