package response

import (
	"encoding/json"
)

type GetVehicle struct {
	ID               int32   `json:"id"`
	IMEI             string  `json:"imei"`
	OID              *int64  `json:"oid,omitempty"`
	Name             *string `json:"name,omitempty"`
	ProviderID       int32   `json:"provider_id"`
	ModerationStatus string  `json:"moderation_status"`
}

func (v GetVehicle) MarshalJSON() ([]byte, error) {
	type out struct {
		ID               int32       `json:"id"`
		IMEI             string      `json:"imei"`
		OID              interface{} `json:"oid,omitempty"`
		Name             interface{} `json:"name,omitempty"`
		ProviderID       int32       `json:"provider_id"`
		ModerationStatus string      `json:"moderation_status"`
	}
	o := out{
		ID:               v.ID,
		IMEI:             v.IMEI,
		ProviderID:       v.ProviderID,
		ModerationStatus: v.ModerationStatus,
	}
	if v.OID != nil {
		o.OID = *v.OID
	}
	if v.Name != nil {
		o.Name = *v.Name
	}
	return json.Marshal(o)
}

type GetVehicles []GetVehicle
