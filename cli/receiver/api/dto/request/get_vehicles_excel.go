package request

import (
	"github.com/daniil11ru/egts/cli/receiver/repository/primary/types"
)

type GetVehiclesExcel struct {
	ProviderID       *int32
	ModerationStatus *types.ModerationStatus
}
