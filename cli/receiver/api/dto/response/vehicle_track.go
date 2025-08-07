package response

import (
	"time"
)

type Location struct {
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	SentAt     time.Time `json:"sent_at"`
	ReceivedAt time.Time `json:"received_at"`
}

type VehicleTrack struct {
	VehicleId int32      `json:"vehicle_id"`
	Locations []Location `json:"locations"`
}
