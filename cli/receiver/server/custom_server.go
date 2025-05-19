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

func NewCustom(addr string, ttl time.Duration, store *storage.Repository) *CustomServer {
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

		isPkgSave := false
		packetIDBytes := make([]byte, 4)

		srResponsesRecord = append(srResponsesRecord, egts.RecordData{
			SubrecordType:   egts.SrRecordResponseType,
			SubrecordLength: 3,
			SubrecordData: &egts.SrResponse{
				ConfirmedRecordNumber: rec.RecordNumber,
				RecordStatus:          egtsPcOk,
			},
		})
		serviceType = rec.SourceServiceType
		log.Debug("Тип сервиса: ", serviceType)

		if rec.ObjectIDFieldExists == "1" {
			client = rec.ObjectIdentifier
		}

		for _, subRec := range rec.RecordDataSet {
			switch subRecData := subRec.SubrecordData.(type) {
			case *egts.SrTermIdentity:
				log.Debug("Разбор подзаписи EGTS_SR_TERM_IDENTITY")
				client = subRecData.TerminalIdentifier
				if bytes, err := createSrResultCode(pkg.PacketIdentifier, egtsPcOk); err == nil {
					srResultCodePkg = bytes
				} else {
					log.Errorf("Ошибка сборки EGTS_SR_RESULT_CODE: %v", err)
				}
			case *egts.SrAuthInfo:
				log.Debug("Разбор подзаписи EGTS_SR_AUTH_INFO")
				if bytes, err := createSrResultCode(pkg.PacketIdentifier, egtsPcOk); err == nil {
					srResultCodePkg = bytes
				} else {
					log.Errorf("Ошибка сборки EGTS_SR_RESULT_CODE: %v", err)
				}
			case *egts.SrResponse:
				log.Debug("Разбор подзаписи EGTS_SR_RESPONSE")
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
			case *egts.SrAdSensorsData:
				log.Debug("Разбор подзаписи EGTS_SR_AD_SENSORS_DATA")
				if subRecData.AnalogSensorFieldExists1 == "1" {
					exportPacket.AnSensors = append(exportPacket.AnSensors, storage.AnSensor{SensorNumber: 1, Value: subRecData.AnalogSensor1})
				}
				if subRecData.AnalogSensorFieldExists2 == "1" {
					exportPacket.AnSensors = append(exportPacket.AnSensors, storage.AnSensor{SensorNumber: 2, Value: subRecData.AnalogSensor2})
				}
				if subRecData.AnalogSensorFieldExists3 == "1" {
					exportPacket.AnSensors = append(exportPacket.AnSensors, storage.AnSensor{SensorNumber: 3, Value: subRecData.AnalogSensor3})
				}
				if subRecData.AnalogSensorFieldExists4 == "1" {
					exportPacket.AnSensors = append(exportPacket.AnSensors, storage.AnSensor{SensorNumber: 4, Value: subRecData.AnalogSensor4})
				}
				if subRecData.AnalogSensorFieldExists5 == "1" {
					exportPacket.AnSensors = append(exportPacket.AnSensors, storage.AnSensor{SensorNumber: 5, Value: subRecData.AnalogSensor5})
				}
				if subRecData.AnalogSensorFieldExists6 == "1" {
					exportPacket.AnSensors = append(exportPacket.AnSensors, storage.AnSensor{SensorNumber: 6, Value: subRecData.AnalogSensor6})
				}
				if subRecData.AnalogSensorFieldExists7 == "1" {
					exportPacket.AnSensors = append(exportPacket.AnSensors, storage.AnSensor{SensorNumber: 7, Value: subRecData.AnalogSensor7})
				}
				if subRecData.AnalogSensorFieldExists8 == "1" {
					exportPacket.AnSensors = append(exportPacket.AnSensors, storage.AnSensor{SensorNumber: 8, Value: subRecData.AnalogSensor8})
				}
			case *egts.SrAbsAnSensData:
				log.Debug("Разбор подзаписи EGTS_SR_ABS_AN_SENS_DATA")
				exportPacket.AnSensors = append(exportPacket.AnSensors, storage.AnSensor{SensorNumber: subRecData.SensorNumber, Value: subRecData.Value})
			case *egts.SrAbsCntrData:
				log.Debug("Разбор подзаписи EGTS_SR_ABS_CNTR_DATA")
				switch subRecData.CounterNumber {
				case 110:
					binary.BigEndian.PutUint32(packetIDBytes, subRecData.CounterValue)
					exportPacket.PacketID = subRecData.CounterValue
				case 111:
					tmp := make([]byte, 4)
					binary.BigEndian.PutUint32(tmp, subRecData.CounterValue)
					packetIDBytes[3] = tmp[3]
					exportPacket.PacketID = binary.LittleEndian.Uint32(packetIDBytes)
				}
			case *egts.SrLiquidLevelSensor:
				log.Debug("Разбор подзаписи EGTS_SR_LIQUID_LEVEL_SENSOR")
				sensorData := storage.LiquidSensor{
					SensorNumber: subRecData.LiquidLevelSensorNumber,
					ErrorFlag:    subRecData.LiquidLevelSensorErrorFlag,
				}
				switch subRecData.LiquidLevelSensorValueUnit {
				case "00", "01":
					sensorData.ValueMm = subRecData.LiquidLevelSensorData
				case "10":
					sensorData.ValueL = subRecData.LiquidLevelSensorData * 10
				}
				exportPacket.LiquidSensors = append(exportPacket.LiquidSensors, sensorData)
			}
		}

		exportPacket.Client = client
		if isPkgSave {
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
