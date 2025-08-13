package filter

import (
	"github.com/daniil11ru/egts/cli/receiver/types"
)

type Vehicles struct {
	ProviderID       *int32
	ModerationStatus *types.ModerationStatus
	IMEI             *string
}
