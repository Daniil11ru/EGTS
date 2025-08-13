package request

import (
	"github.com/daniil11ru/egts/cli/receiver/types"
)

type GetVehiclesExcel struct {
	ProviderID       *int32
	ModerationStatus *types.ModerationStatus
}
