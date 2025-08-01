package model

import "time"

type LatestLocation struct {
	VehicleId      int32     `json:"vehicle_id"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Altitude       int16     `json:"altitude"`
	Direction      int16     `json:"direction"`
	Speed          int16     `json:"speed"`
	SatelliteCount int8      `json:"satellite_count"`
	SentAt         time.Time `json:"sent_at"`
	ReceivedAt     time.Time `json:"received_at"`
}
