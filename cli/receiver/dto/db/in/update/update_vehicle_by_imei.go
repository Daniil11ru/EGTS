package update

import "github.com/daniil11ru/egts/cli/receiver/dto/other"

type VehicleByImei struct {
	Name             *string                 `json:"name"`
	ModerationStatus *other.ModerationStatus `json:"moderation_status"`
}
