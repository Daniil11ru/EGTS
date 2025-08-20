package response

import "encoding/json"

type Location struct {
	OID            int64   `json:"oid"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Altitude       *int64  `json:"altitude,omitempty"`
	Direction      *int16  `json:"direction,omitempty"`
	Speed          *int32  `json:"speed,omitempty"`
	SatelliteCount *int16  `json:"satellite_count,omitempty"`
	SentAt         *string `json:"sent_at,omitempty"`
	ReceivedAt     string  `json:"received_at"`
}

func (l Location) MarshalJSON() ([]byte, error) {
	type out struct {
		OID            int64       `json:"oid"`
		Latitude       float64     `json:"latitude"`
		Longitude      float64     `json:"longitude"`
		Altitude       interface{} `json:"altitude,omitempty"`
		Direction      interface{} `json:"direction,omitempty"`
		Speed          interface{} `json:"speed,omitempty"`
		SatelliteCount interface{} `json:"satellite_count,omitempty"`
		SentAt         interface{} `json:"sent_at,omitempty"`
		ReceivedAt     string      `json:"received_at"`
	}
	o := out{
		OID:        l.OID,
		Latitude:   l.Latitude,
		Longitude:  l.Longitude,
		ReceivedAt: l.ReceivedAt,
	}
	if l.Altitude != nil {
		o.Altitude = *l.Altitude
	}
	if l.Direction != nil {
		o.Direction = *l.Direction
	}
	if l.Speed != nil {
		o.Speed = *l.Speed
	}
	if l.SatelliteCount != nil {
		o.SatelliteCount = *l.SatelliteCount
	}
	if l.SentAt != nil {
		o.SentAt = *l.SentAt
	}
	return json.Marshal(o)
}

type VehicleTrack struct {
	VehicleId int32      `json:"vehicle_id"`
	Locations []Location `json:"locations"`
}

type GetLocations []VehicleTrack
