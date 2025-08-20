package request

import "github.com/daniil11ru/egts/cli/receiver/dto/other"

type UpdateVehicle struct {
	IMEI             *string                 `json:"imei"`
	Name             *string                 `json:"name"`
	ModerationStatus *other.ModerationStatus `json:"moderation_status"`
}
