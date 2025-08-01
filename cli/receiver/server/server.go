package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/daniil11ru/egts/cli/receiver/domain"
	packet "github.com/daniil11ru/egts/cli/receiver/repository/util"
	"github.com/daniil11ru/egts/libs/egts"
	log "github.com/sirupsen/logrus"
)

const (
	egtsPcOk         = 0
	egtsPcSrvcDenied = 0x95
	egtsPcUnsType    = 133
	headerLen        = 10
)

type Server struct {
	Address        string
	TTL            time.Duration
	SavePacket     *domain.SavePacket
	Listener       net.Listener
	GetIPWhiteList domain.GetIPWhiteList
}

func NewServer(addr string, ttl time.Duration, savePacket *domain.SavePacket, getIPWhiteList domain.GetIPWhiteList) *Server {
	return &Server{Address: addr, TTL: ttl, SavePacket: savePacket, GetIPWhiteList: getIPWhiteList}
}

func extractIp(ipAndPort string) (string, error) {
	var re = regexp.MustCompile(`^((?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)){3}):(\d{1,5})$`)
	matches := re.FindStringSubmatch(ipAndPort)
	if matches == nil {
		return "", fmt.Errorf("не удалось получить IP-адрес")
	}

	return matches[1], nil
}

func isInWhiteList(ip string, whiteList []string) bool {
	isInWhiteList := false

ipCheckLoop:
	for _, entry := range whiteList {
		if entry == ip {
			isInWhiteList = true
			break
		}

		if strings.Contains(entry, "*") {
			partsEntry := strings.Split(entry, ".")
			partsIP := strings.Split(ip, ".")

			if partsEntry[len(partsEntry)-1] != "*" {
				continue
			}

			prefixEntry := partsEntry[:len(partsEntry)-1]
			if len(prefixEntry) > len(partsIP) {
				continue
			}

			for i := 0; i < len(prefixEntry); i++ {
				if prefixEntry[i] != partsIP[i] {
					continue ipCheckLoop
				}
			}

			isInWhiteList = true
			break
		}
	}

	return isInWhiteList
}

func (server *Server) Run() error {
	var err error
	server.Listener, err = net.Listen("tcp", server.Address)
	if err != nil {
		return fmt.Errorf("не удалось открыть соединение: %w", err)
	}
	defer server.Listener.Close()

	log.WithField("addr", server.Address).Info("Запущен сервер")

	whiteList, err := server.GetIPWhiteList.Run()
	if err != nil || len(whiteList) == 0 {
		return fmt.Errorf("не удалось получить белый список IP: %w", err)
	}
	log.Debug("Белый список IP: ", whiteList)

	for {
		conn, err := server.Listener.Accept()
		if err != nil {
			log.WithField("err", err).Error("Ошибка соединения")
			continue
		}

		ip, getIpFromIPAndPortErr := extractIp(conn.RemoteAddr().String())
		if getIpFromIPAndPortErr != nil {
			log.Warn("Адрес отправителя не является IP-адресом")
			conn.Close()
			continue
		}

		if !isInWhiteList(ip, whiteList) {
			log.Warnf("IP %s не находится в белом списке", ip)
			conn.Close()
			continue
		}

		server.handleConnection(conn)
	}
}

func (s *Server) handleConnection(connection net.Conn) {
	defer connection.Close()

	log.WithField("ip", connection.RemoteAddr()).Info("Установлено соединение")

	for {
		packet, err := s.readPacket(connection)
		if err != nil {
			return
		}

		pkg, receivedTimestamp, resultCode, err := s.decodePacket(packet)
		if err != nil {
			s.sendDecodeError(connection, pkg.PacketIdentifier, resultCode)
			continue
		}

		switch pkg.PacketType {
		case egts.PtAppdataPacket:
			if err := s.handleAppData(connection, pkg, receivedTimestamp, resultCode); err != nil {
				continue
			}
		case egts.PtResponsePacket:
			log.Debug("Тип пакета EGTS_PT_RESPONSE")
		}
	}
}

func (s *Server) readPacket(conn net.Conn) ([]byte, error) {
	if s.TTL > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.TTL))
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

	if s.TTL > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.TTL))
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

func (s *Server) decodePacket(packet []byte) (*egts.Package, int64, uint8, error) {
	pkg := egts.Package{}
	receivedTimestamp := time.Now().UTC().Unix()
	resultCode, err := pkg.Decode(packet)
	return &pkg, receivedTimestamp, resultCode, err
}

func (s *Server) sendDecodeError(conn net.Conn, packetIdentifier uint16, resultCode uint8) {
	resp, err := createPtResponse(packetIdentifier, resultCode, 0, nil)
	if err != nil {
		log.WithField("err", err).Error("Ошибка сборки ответа EGTS_PT_RESPONSE с ошибкой")
		return
	}
	_, _ = conn.Write(resp)
}

func (s *Server) handleAppData(conn net.Conn, pkg *egts.Package, receivedTimestamp int64, resultCode uint8) error {
	var (
		srResponsesRecord egts.RecordDataSet
		srResultCodePkg   []byte
		serviceType       uint8
		client            uint32
	)

	for _, rec := range *pkg.ServicesFrameData.(*egts.ServiceDataSet) {
		exportPacket := packet.NavigationRecord{}

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
				exportPacket.SentTimestamp = subRecData.NavigationTime.Unix()
				exportPacket.ReceivedTimestamp = receivedTimestamp
				exportPacket.Latitude = subRecData.Latitude
				exportPacket.Longitude = subRecData.Longitude
				exportPacket.Altitude = subRecData.Altitude
				exportPacket.Speed = subRecData.Speed
				exportPacket.Direction = subRecData.Direction
			case *egts.SrExtPosData:
				log.Debug("Разбор подзаписи EGTS_SR_EXT_POS_DATA")
				exportPacket.SatelliteCount = subRecData.Satellites
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

		exportPacket.OID = client
		if isPkgSave && recStatus == egtsPcOk {
			IP, getIPFromIPAndPortErr := extractIp(conn.RemoteAddr().String())
			if getIPFromIPAndPortErr != nil {
				log.Warn("Адрес отправителя не является IP-адресом")
			} else {
				pkt := exportPacket
				go func() {
					if err := s.SavePacket.Run(&pkt, IP); err != nil {
						log.Warnf("Телематические данные не были сохранены: %s", err)
					}
				}()
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

func createPtResponse(pid uint16, resultCode, serviceType uint8, srResponses egts.RecordDataSet) ([]byte, error) {
	respSection := egts.PtResponse{
		ResponsePacketID: pid,
		ProcessingResult: resultCode,
	}

	if srResponses != nil {
		respSection.SDR = &egts.ServiceDataSet{
			egts.ServiceDataRecord{
				RecordLength:             srResponses.Length(),
				RecordNumber:             1,
				SourceServiceOnDevice:    "0",
				RecipientServiceOnDevice: "0",
				Group:                    "1",
				RecordProcessingPriority: "00",
				TimeFieldExists:          "0",
				EventIDFieldExists:       "0",
				ObjectIDFieldExists:      "0",
				SourceServiceType:        serviceType,
				RecipientServiceType:     serviceType,
				RecordDataSet:            srResponses,
			},
		}
	}

	respPkg := egts.Package{
		ProtocolVersion:   1,
		SecurityKeyID:     0,
		Prefix:            "00",
		Route:             "0",
		EncryptionAlg:     "00",
		Compression:       "0",
		Priority:          "00",
		HeaderLength:      11,
		HeaderEncoding:    0,
		FrameDataLength:   respSection.Length(),
		PacketIdentifier:  pid + 1,
		PacketType:        egts.PtResponsePacket,
		ServicesFrameData: &respSection,
	}

	return respPkg.Encode()
}
