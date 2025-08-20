package filter

import (
	"github.com/daniil11ru/egts/cli/receiver/dto/other"
)

type Vehicles struct {
	ProviderID       *int32
	ModerationStatus *other.ModerationStatus
	IMEI             *string
	OID              *int64
}
