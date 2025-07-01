package util

import (
	"encoding/json"
)

type NavigationRecord struct {
	OID               uint32  `json:"oid"`
	SentTimestamp     int64   `json:"sent_unix_time"`
	ReceivedTimestamp int64   `json:"received_unix_time"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	Altitude          uint32  `json:"altitude"`
	Speed             uint16  `json:"speed"`
	SatelliteCount    uint8   `json:"satellite_count"`
	Direction         uint8   `json:"direction"`
}

func (eep *NavigationRecord) ToBytes() ([]byte, error) {
	return json.Marshal(eep)
}
