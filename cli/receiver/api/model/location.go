package model

import "time"

type Location struct {
	VehicleId  int32     `json:"vehicle_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	SentAt     time.Time `json:"sent_at"`
	ReceivedAt time.Time `json:"received_at"`
}
