package request

import "github.com/daniil11ru/egts/cli/receiver/repository/primary/types"

type UpdateVehicle struct {
	IMEI             string                  `json:"imei"`
	Name             *string                 `json:"name"`
	ModerationStatus *types.ModerationStatus `json:"moderation_status"`
}
