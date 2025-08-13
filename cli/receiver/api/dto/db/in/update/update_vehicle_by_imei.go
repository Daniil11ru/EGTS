package update

import "github.com/daniil11ru/egts/cli/receiver/types"

type VehicleByImei struct {
	Name             *string                 `json:"name"`
	ModerationStatus *types.ModerationStatus `json:"moderation_status"`
}
