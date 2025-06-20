package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/daniil11ru/egts/libs/egts"
)

/*
EGTS packet generator.

Util create egts packet  from setting parameters.

Usage:
  -pid int
    	Packet identifier (require)
  -oid int
    	Client identifier (require)
  -time string
    	Timestamp in RFC 3339 format (require)
  -lat float
    	Latitude
  -liquid int
    	Liquid level for first sensor
  -lon float
    	Longitude
  -server string
    	Egts server address in format <ip>:<port> (default "localhost:5555")
  -timeout int
    	Ack waiting time in seconds, Default: 5

Example

```
./packet-gen --pid 1 --oid 12 --time 2021-12-16T09:12:00Z --lat 45 --lon 60.344 --server localhost:5555
```

Created by Igor Kuznetsov
*/

func main() {

	pid := 0
	oid := 0
	ts := ""
	liqLvl := 0
	lat := 0.0
	lon := 0.0
	server := ""
	ackTimeout := 0
	pktType := ""

	flag.IntVar(&pid, "pid", 0, "Идентификатор пакета (обязательно)")
	flag.IntVar(&oid, "oid", 0, "Идентификатор клиента (обязательно)")
	flag.StringVar(&ts, "time", "", "Метка времени в формате RFC 3339 (обязательно)")
	flag.IntVar(&liqLvl, "liquid", 0, "Уровень жидкости для первого датчика")
	flag.Float64Var(&lat, "lat", 0, "Широта")
	flag.Float64Var(&lon, "lon", 0, "Долгота")
	flag.IntVar(&ackTimeout, "timeout", 0, "Время ожидания подтверждения в секундах, по умолчанию 5")
	flag.StringVar(&server, "server", "localhost:5555", "Адрес EGTS-сервера в формате <ip>:<port>")
	flag.StringVar(&pktType, "type", "data", "Тип отправляемого пакета: auth, tele, mixed")

	flag.Parse()

	if pid == 0 {
		fmt.Println("Требуется идентификатор пакета, смотрите помощь (-h)")
		os.Exit(1)
	}

	if oid == 0 {
		fmt.Println("Требуется идентификатор клиента, смотрите помощь (-h)")
		os.Exit(1)
	}

	if ts == "" {
		fmt.Println("Требуется метка времени, смотрите помощь (-h)")
		os.Exit(1)
	}
	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		fmt.Println("Ошибка парсинга метки времени: ", timestamp)
		os.Exit(1)
	}

	authPkg := egts.Package{
		ProtocolVersion:  1,
		SecurityKeyID:    0,
		Prefix:           "00",
		Route:            "0",
		EncryptionAlg:    "00",
		Compression:      "0",
		Priority:         "10",
		HeaderLength:     11,
		HeaderEncoding:   0,
		PacketIdentifier: uint16(pid),
		PacketType:       1,
		ServicesFrameData: &egts.ServiceDataSet{
			egts.ServiceDataRecord{
				RecordNumber:             1,
				SourceServiceOnDevice:    "0",
				RecipientServiceOnDevice: "0",
				Group:                    "0",
				RecordProcessingPriority: "10",
				TimeFieldExists:          "0",
				EventIDFieldExists:       "1",
				ObjectIDFieldExists:      "1",
				EventIdentifier:          3436,
				ObjectIdentifier:         uint32(oid),
				SourceServiceType:        1,
				RecipientServiceType:     1,
				RecordDataSet: egts.RecordDataSet{
					egts.RecordData{
						SubrecordType: 7,
						SubrecordData: &egts.SrAuthInfo{
							UserName:     "test",
							UserPassword: "test",
						},
					},
				},
			},
		},
	}

	mixedDataPkg := egts.Package{
		ProtocolVersion:  1,
		SecurityKeyID:    0,
		Prefix:           "00",
		Route:            "0",
		EncryptionAlg:    "00",
		Compression:      "0",
		Priority:         "10",
		HeaderLength:     11,
		HeaderEncoding:   0,
		PacketIdentifier: uint16(pid),
		PacketType:       1,
		ServicesFrameData: &egts.ServiceDataSet{
			egts.ServiceDataRecord{
				RecordNumber:             1,
				SourceServiceOnDevice:    "0",
				RecipientServiceOnDevice: "0",
				Group:                    "0",
				RecordProcessingPriority: "10",
				TimeFieldExists:          "0",
				EventIDFieldExists:       "1",
				ObjectIDFieldExists:      "1",
				EventIdentifier:          3436,
				ObjectIdentifier:         uint32(oid),
				SourceServiceType:        2,
				RecipientServiceType:     2,
				RecordDataSet: egts.RecordDataSet{
					egts.RecordData{
						SubrecordType: 16,
						SubrecordData: &egts.SrPosData{
							NavigationTime:      time.Date(2021, time.February, 20, 0, 30, 40, 0, time.UTC),
							Latitude:            lat,
							Longitude:           lon,
							ALTE:                "1",
							LOHS:                "0",
							LAHS:                "0",
							MV:                  "1",
							BB:                  "1",
							CS:                  "0",
							FIX:                 "1",
							VLD:                 "1",
							DirectionHighestBit: 0,
							AltitudeSign:        0,
							Speed:               34,
							Direction:           172,
							Odometer:            191,
							DigitalInputs:       144,
							Source:              0,
							Altitude:            30,
						},
					},
					egts.RecordData{
						SubrecordType: 27,
						SubrecordData: &egts.SrLiquidLevelSensor{
							LiquidLevelSensorErrorFlag: "1",
							LiquidLevelSensorValueUnit: "00",
							RawDataFlag:                "0",
							LiquidLevelSensorNumber:    1,
							ModuleAddress:              uint16(1),
							LiquidLevelSensorData:      uint32(liqLvl),
						},
					},
				},
			},
		},
	}

	telematicDataPkg := egts.Package{
		ProtocolVersion:  1,
		SecurityKeyID:    0,
		Prefix:           "00",
		Route:            "0",
		EncryptionAlg:    "00",
		Compression:      "0",
		Priority:         "10",
		HeaderLength:     11,
		HeaderEncoding:   0,
		PacketIdentifier: uint16(pid),
		PacketType:       1,
		ServicesFrameData: &egts.ServiceDataSet{
			egts.ServiceDataRecord{
				RecordNumber:             1,
				SourceServiceOnDevice:    "0",
				RecipientServiceOnDevice: "0",
				Group:                    "0",
				RecordProcessingPriority: "10",
				TimeFieldExists:          "0",
				EventIDFieldExists:       "1",
				ObjectIDFieldExists:      "1",
				EventIdentifier:          3436,
				ObjectIdentifier:         uint32(oid),
				SourceServiceType:        2,
				RecipientServiceType:     2,
				RecordDataSet: egts.RecordDataSet{
					egts.RecordData{
						SubrecordType: 16,
						SubrecordData: &egts.SrPosData{
							NavigationTime:      time.Date(2021, time.February, 20, 0, 30, 40, 0, time.UTC),
							Latitude:            lat,
							Longitude:           lon,
							ALTE:                "1",
							LOHS:                "0",
							LAHS:                "0",
							MV:                  "1",
							BB:                  "1",
							CS:                  "0",
							FIX:                 "1",
							VLD:                 "1",
							DirectionHighestBit: 0,
							AltitudeSign:        0,
							Speed:               34,
							Direction:           172,
							Odometer:            191,
							DigitalInputs:       144,
							Source:              0,
							Altitude:            30,
						},
					},
				},
			},
		},
	}

	var pkg egts.Package
	switch pktType {
	case "auth":
		pkg = authPkg
	case "mixed":
		pkg = mixedDataPkg
	case "tele":
		pkg = telematicDataPkg
	default:
		fmt.Println("Неверный тип пакета, используйте auth, tele или mixed в качестве значения параметра -type")
		os.Exit(1)
	}

	sendBytes, err := pkg.Encode()
	if err != nil {
		fmt.Println("Ошибка кодирования сообщения: ", err)
		os.Exit(1)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		fmt.Println("Ошибка преобразования адреса: ", err)
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Ошибка соединения: ", err)
		os.Exit(1)
	}

	_, err = conn.Write(sendBytes)
	if err != nil {
		fmt.Println("Ошибка записи на сервер: ", err)
		os.Exit(1)
	}

	ackBuf := make([]byte, 1024)

	_ = conn.SetReadDeadline(time.Now().Add(time.Duration(ackTimeout) * time.Second))
	ackLen, err := conn.Read(ackBuf)
	if err != nil {
		fmt.Println("Ошибка чтения с сервера: ", err)
		os.Exit(1)
	}

	ackPacket := egts.Package{}
	_, err = ackPacket.Decode(ackBuf[:ackLen])
	if err != nil {
		fmt.Println("Ошибка разбора ACK-пакета: ", err)
		os.Exit(1)
	}

	ack, ok := ackPacket.ServicesFrameData.(*egts.PtResponse)
	if !ok {
		fmt.Println("Полученный пакет не является EGTS ACK")
		os.Exit(1)
	}

	if ack.ResponsePacketID != pkg.PacketIdentifier {
		fmt.Printf("Некорректный идентификатор пакета-ответа: %d (фактический) != %d (ожидаемый)",
			ack.ResponsePacketID, pkg.PacketIdentifier)
		os.Exit(1)
	}

	if ack.ProcessingResult != 0 {
		fmt.Printf("Некорректный результат обработки: %d (фактический) != 0 (ожидаемый)", ack.ProcessingResult)
		os.Exit(1)
	}

	for _, rec := range *ack.SDR.(*egts.ServiceDataSet) {
		for _, subRec := range rec.RecordDataSet {
			if subRec.SubrecordType == egts.SrRecordResponseType {
				if response, ok := subRec.SubrecordData.(*egts.SrResponse); ok {
					fmt.Printf("Код ответа: %v\n", response.RecordStatus)
				}
			}
		}
	}

	fmt.Println("Пакет отправлен и корректно обработан сервером")
	os.Exit(0)
}
