package server

import (
	"encoding/binary"
	"io"
	"net"
	"time"

	domain "github.com/daniil11ru/egts/cli/receiver/domain"
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
	addr        string
	ttl         time.Duration
	savePackage *domain.SavePackage
	l           net.Listener
}

func (s *Server) Run() {
	var err error

	s.l, err = net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("Не удалось открыть соединение: %v", err)
	}
	defer s.l.Close()

	log.Infof("Запущен сервер %s", s.addr)
	for {
		conn, err := s.l.Accept()
		if err != nil {
			log.WithField("err", err).Errorf("Ошибка соединения")
		} else {
			go s.handleConn(conn)
		}
	}
}

func (s *Server) Stop() error {
	if s.l != nil {
		return s.l.Close()
	}

	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	var (
		isPkgSave         bool
		srResultCodePkg   []byte
		serviceType       uint8
		srResponsesRecord egts.RecordDataSet
		recvPacket        []byte
		client            uint32
	)

	log.WithField("ip", conn.RemoteAddr()).Info("Установлено соединение")

	log.Debug("TTL: ", s.ttl)

	for {
	Received:
		serviceType = 0
		srResponsesRecord = nil
		srResultCodePkg = nil
		recvPacket = nil

		if s.ttl > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(s.ttl))
		} else {
			_ = conn.SetReadDeadline(time.Time{})
		}

		// считываем заголовок пакета
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
			return
		}

		// Закрываем соединение, если пакет не формата EGTS
		if headerBuf[0] != 0x01 {
			log.WithField("ip", conn.RemoteAddr()).Warn("Пакет не соответствует формату EGTS")
			_ = conn.SetDeadline(time.Time{})
			return
		}

		// вычисляем длину пакета, равную длине заголовка (HL) + длина тела (FDL) + CRC пакета 2 байта если есть FDL из приказа минтранса №285
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

		// получаем концовку ЕГТС пакета
		buf := make([]byte, pkgLen-headerLen)
		if _, err := io.ReadFull(conn, buf); err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				log.WithField("ip", conn.RemoteAddr()).Warn("Таймаут чтения")
			} else {
				log.WithField("err", err).Error("Ошибка при получении тела пакета")
			}
			_ = conn.SetDeadline(time.Time{})
			return
		}

		// формируем полный пакет
		recvPacket = append(headerBuf, buf...)
		_ = conn.SetReadDeadline(time.Time{})

		log.WithField("packet", recvPacket).Debug("Принят пакет")
		pkg := egts.Package{}
		receivedTimestamp := time.Now().UTC().Unix()
		resultCode, err := pkg.Decode(recvPacket)
		if err != nil {
			log.WithField("err", err).Error("Ошибка расшифровки пакета")

			resp, err := createPtResponse(pkg.PacketIdentifier, resultCode, serviceType, nil)
			if err != nil {
				log.WithField("err", err).Error("Ошибка сборки ответа EGTS_PT_RESPONSE с ошибкой")
				goto Received
			}
			_, _ = conn.Write(resp)

			goto Received
		}

		switch pkg.PacketType {
		case egts.PtAppdataPacket:
			log.Debug("Тип пакета EGTS_PT_APPDATA")

			for _, rec := range *pkg.ServicesFrameData.(*egts.ServiceDataSet) {
				exportPacket := packet.NavigationRecord{
					PacketID: uint32(pkg.PacketIdentifier),
				}

				isPkgSave = false
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
				log.Info("Тип сервиса ", serviceType)

				// если в секции с данными есть oid, то обновляем его
				if rec.ObjectIDFieldExists == "1" {
					client = rec.ObjectIdentifier
				}

				for _, subRec := range rec.RecordDataSet {
					switch subRecData := subRec.SubrecordData.(type) {
					case *egts.SrTermIdentity:
						log.Debug("Разбор подзаписи EGTS_SR_TERM_IDENTITY")

						// на случай если секция с данными не содержит oid
						client = subRecData.TerminalIdentifier

						if srResultCodePkg, err = createSrResultCode(pkg.PacketIdentifier, egtsPcOk); err != nil {
							log.Errorf("Ошибка сборки EGTS_SR_RESULT_CODE: %v", err)
						}
					case *egts.SrAuthInfo:
						log.Debug("Разбор подзаписи EGTS_SR_AUTH_INFO")
						if srResultCodePkg, err = createSrResultCode(pkg.PacketIdentifier, egtsPcOk); err != nil {
							log.Errorf("Ошибка сборки EGTS_SR_RESULT_CODE: %v", err)
						}
					case *egts.SrResponse:
						log.Debugf("Разбор подзаписи EGTS_SR_RESPONSE")
						goto Received
					case *egts.SrPosData:
						log.Debugf("Разбор подзаписи EGTS_SR_POS_DATA")
						isPkgSave = true

						exportPacket.SentTimestamp = subRecData.NavigationTime.Unix()
						exportPacket.ReceivedTimestamp = receivedTimestamp
						exportPacket.Latitude = subRecData.Latitude
						exportPacket.Longitude = subRecData.Longitude
						exportPacket.Speed = subRecData.Speed
						exportPacket.Direction = subRecData.Direction
					case *egts.SrExtPosData:
						log.Debug("Разбор подзаписи EGTS_SR_EXT_POS_DATA")
						exportPacket.SatelliteCount = subRecData.Satellites
						exportPacket.PDOP = subRecData.PositionDilutionOfPrecision
						exportPacket.HDOP = subRecData.HorizontalDilutionOfPrecision
						exportPacket.VDOP = subRecData.VerticalDilutionOfPrecision
						exportPacket.NavigationSystem = subRecData.NavigationSystem

					case *egts.SrAdSensorsData:
						log.Debug("Встречена подзапись EGTS_SR_AD_SENSORS_DATA")
					case *egts.SrAbsAnSensData:
						log.Debug("Встречена подзапись EGTS_SR_ABS_AN_SENS_DATA")
					case *egts.SrAbsCntrData:
						log.Debug("Разбор подзаписи EGTS_SR_ABS_CNTR_DATA")

						switch subRecData.CounterNumber {
						case 110:
							// Три младших байта номера передаваемой записи (идет вместе с каждой POS_DATA).
							binary.BigEndian.PutUint32(packetIDBytes, subRecData.CounterValue)
							exportPacket.PacketID = subRecData.CounterValue
						case 111:
							// один старший байт номера передаваемой записи (идет вместе с каждой POS_DATA).
							tmpBuf := make([]byte, 4)
							binary.BigEndian.PutUint32(tmpBuf, subRecData.CounterValue)

							if len(packetIDBytes) == 4 {
								packetIDBytes[3] = tmpBuf[3]
							} else {
								packetIDBytes = tmpBuf
							}

							exportPacket.PacketID = binary.LittleEndian.Uint32(packetIDBytes)
						}
					case *egts.SrLiquidLevelSensor:
						log.Debug("Встречена подзапись EGTS_SR_LIQUID_LEVEL_SENSOR")
					}
				}

				exportPacket.OID = client
				if isPkgSave {
					pkt := exportPacket
					ip := conn.RemoteAddr().String()
					go func() {
						if err := s.savePackage.Run(&pkt, ip); err != nil {
							log.Warnf("Телематические данные не были сохранены: %v", err)
						}
					}()
				}
			}

			resp, err := createPtResponse(pkg.PacketIdentifier, resultCode, serviceType, srResponsesRecord)
			if err != nil {
				log.WithField("err", err).Error("Ошибка сборки ответа")
				goto Received
			}
			_, _ = conn.Write(resp)

			log.WithField("packet", resp).Debug("Отправлен пакет EGTS_PT_RESPONSE")

			if len(srResultCodePkg) > 0 {
				_, _ = conn.Write(srResultCodePkg)
				log.WithField("packet", resp).Debug("Отправлен пакет EGTS_SR_RESULT_CODE")
			}
		case egts.PtResponsePacket:
			log.Debug("Тип пакета EGTS_PT_RESPONSE")
		}

	}
}

func New(srvAddress string, ttl time.Duration, savePackage *domain.SavePackage) Server {
	return Server{
		addr:        srvAddress,
		ttl:         ttl,
		savePackage: savePackage,
	}
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

func createSrResultCode(pid uint16, resultCode uint8) ([]byte, error) {
	rds := egts.RecordDataSet{
		egts.RecordData{
			SubrecordType:   egts.SrResultCodeType,
			SubrecordLength: uint16(1),
			SubrecordData: &egts.SrResultCode{
				ResultCode: resultCode,
			},
		},
	}

	sfd := egts.ServiceDataSet{
		egts.ServiceDataRecord{
			RecordLength:             rds.Length(),
			RecordNumber:             1,
			SourceServiceOnDevice:    "0",
			RecipientServiceOnDevice: "0",
			Group:                    "1",
			RecordProcessingPriority: "00",
			TimeFieldExists:          "0",
			EventIDFieldExists:       "0",
			ObjectIDFieldExists:      "0",
			SourceServiceType:        egts.AuthService,
			RecipientServiceType:     egts.AuthService,
			RecordDataSet:            rds,
		},
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
		FrameDataLength:   sfd.Length(),
		PacketIdentifier:  pid + 1,
		PacketType:        egts.PtResponsePacket,
		ServicesFrameData: &sfd,
	}

	return respPkg.Encode()
}
