package update

import (
	"github.com/daniil11ru/egts/cli/receiver/dto/other"
)

type VehicleById struct {
	IMEI             *string                 `json:"imei"`
	OID              *int64                  `json:"oid"`
	Name             *string                 `json:"name"`
	ModerationStatus *other.ModerationStatus `json:"moderation_status"`
}
