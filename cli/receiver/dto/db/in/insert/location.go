package insert

import "time"

type Location struct {
	VehicleId      int32      `json:"vehicle_id"`
	OID            int64      `json:"oid"`
	Latitude       float64    `json:"latitude"`
	Longitude      float64    `json:"longitude"`
	Altitude       *int64     `json:"altitude"`
	Direction      *int16     `json:"direction"`
	Speed          *int32     `json:"speed"`
	SatelliteCount *int16     `json:"satellite_count"`
	SentAt         *time.Time `json:"sent_at"`
	ReceivedAt     time.Time  `json:"received_at"`
}
