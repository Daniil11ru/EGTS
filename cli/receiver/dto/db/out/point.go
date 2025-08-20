package out

import (
	"fmt"
	"math"
)

type Point struct {
	LocationId int32   `json:"location_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Altitude   *int64  `json:"altitude"`
}

func (p *Point) EqualsTo(position *Point, accuracy_meters float64) (bool, error) {
	if p == nil || position == nil || p.Altitude == nil || position.Altitude == nil {
		return false, fmt.Errorf("точки и их составные части не должны быть nil")
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

	altDiff := float64(*p.Altitude) - float64(*position.Altitude)
	if altDiff < 0 {
		altDiff = -altDiff
	}
	dist := math.Sqrt(horiz*horiz + altDiff*altDiff)
	return dist <= accuracy_meters, nil
}

func (p *Point) EqualsHorizontallyTo(position *Point, accuracy_meters float64) (bool, error) {
	if p == nil || position == nil {
		return false, fmt.Errorf("точки не должны быть nil")
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
	return dist <= accuracy_meters, nil
}
