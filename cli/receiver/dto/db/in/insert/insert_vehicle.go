package insert

import "github.com/daniil11ru/egts/cli/receiver/dto/other"

type Vehicle struct {
	ID               int32                  `json:"id"`
	IMEI             string                 `json:"imei"`
	OID              *int64                 `json:"oid"`
	Name             *string                `json:"name"`
	ProviderID       int32                  `json:"provider_id"`
	ModerationStatus other.ModerationStatus `json:"moderation_status"`
}
