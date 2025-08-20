package filter

import "time"

type Locations struct {
	VehicleId      *int32
	SentBefore     *time.Time
	SentAfter      *time.Time
	ReceivedBefore *time.Time
	ReceivedAfter  *time.Time
	LocationsLimit int64
}
