package types

import "math"

type Position2D struct {
	Latitude  float64
	Longitude float64
}

type Position3D struct {
	Latitude  float64
	Longitude float64
	Altitude  uint32
}

func (p *Position3D) EqualsTo(position *Position3D, accuracy_meters float64) bool {
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

func (p *Position3D) EqualsHorizontallyTo(position *Position3D, accuracy_meters float64) bool {
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
