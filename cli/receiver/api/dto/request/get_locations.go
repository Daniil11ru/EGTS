package request

import "time"

type GetLocations struct {
	VehicleID      *int32
	SentBefore     *time.Time
	SentAfter      *time.Time
	ReceivedBefore *time.Time
	ReceivedAfter  *time.Time
}
