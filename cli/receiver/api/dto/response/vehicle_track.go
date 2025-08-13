package response

type Location struct {
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Altitude       int16   `json:"altitude"`
	Direction      int16   `json:"direction"`
	Speed          int16   `json:"speed"`
	SatelliteCount int8    `json:"satellite_count"`
	SentAt         string  `json:"sent_at"`
	ReceivedAt     string  `json:"received_at"`
}

type VehicleTrack struct {
	VehicleId int32      `json:"vehicle_id"`
	Locations []Location `json:"locations"`
}
