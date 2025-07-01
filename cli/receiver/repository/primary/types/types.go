package types

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
)

type ModerationStatus string

const (
	ModerationStatusPending  ModerationStatus = "pending"
	ModerationStatusApproved ModerationStatus = "approved"
	ModerationStatusRejected ModerationStatus = "rejected"
)

func (ms *ModerationStatus) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan ModerationStatus from %T", value)
	}
	*ms = ModerationStatus(string(b))
	return nil
}

func (ms ModerationStatus) Value() (driver.Value, error) {
	return string(ms), nil
}

type Vehicle struct {
	ID               int32
	IMEI             int64
	Name             sql.NullString
	ProviderID       int32
	ModerationStatus ModerationStatus
}

type Provider struct {
	ID   int32
	Name string
	IP   string
}

type Position struct {
	Latitude  float64
	Longitude float64
	Altitude  uint32
}

func (p *Position) EqualsTo(position *Position, accuracy_meters float64) bool {
	if p == nil || position == nil {
		return false
	}
	const R = 6371000.0
	lat1 := p.Latitude * math.Pi / 180
	lat2 := position.Latitude * math.Pi / 180
	dLat := (position.Latitude - p.Latitude) * math.Pi / 180
	dLon := (position.Longitude - p.Longitude) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	horiz := R * c

	altDiff := float64(p.Altitude) - float64(position.Altitude)
	if altDiff < 0 {
		altDiff = -altDiff
	}
	dist := math.Sqrt(horiz*horiz + altDiff*altDiff)
	return dist <= accuracy_meters
}

func (p *Position) EqualsHorizontallyTo(position *Position, accuracy_meters float64) bool {
	if p == nil || position == nil {
		return false
	}
	const R = 6371000.0
	lat1 := p.Latitude * math.Pi / 180
	lat2 := position.Latitude * math.Pi / 180
	dLat := (position.Latitude - p.Latitude) * math.Pi / 180
	dLon := (position.Longitude - p.Longitude) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	dist := R * c
	return dist <= accuracy_meters
}
