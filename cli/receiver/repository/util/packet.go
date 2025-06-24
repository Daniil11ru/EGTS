package util

import (
	"encoding/json"
)

type NavigationRecord struct {
	OID               uint32  `json:"oid"`
	PacketID          uint32  `json:"packet_id"`
	SentTimestamp     int64   `json:"sent_unix_time"`
	ReceivedTimestamp int64   `json:"received_unix_time"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	Altitude          uint32  `json:"altitude"`
	Speed             uint16  `json:"speed"`
	PDOP              uint16  `json:"pdop"`
	HDOP              uint16  `json:"hdop"`
	VDOP              uint16  `json:"vdop"`
	SatelliteCount    uint8   `json:"satellite_count"`
	NavigationSystem  uint16  `json:"navigation_system"`
	Direction         uint8   `json:"direction"`
}

func (eep *NavigationRecord) ToBytes() ([]byte, error) {
	return json.Marshal(eep)
}
