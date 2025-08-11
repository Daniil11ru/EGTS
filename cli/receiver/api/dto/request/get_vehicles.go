package request

import (
	"github.com/daniil11ru/egts/cli/receiver/repository/primary/types"
)

type GetVehicles struct {
	ProviderID       *int32
	ModerationStatus *types.ModerationStatus
	IMEI             *int64
}
