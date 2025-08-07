package response

import (
	"database/sql"
	"encoding/json"
)

type Vehicle struct {
	ID               int32          `json:"id"`
	IMEI             string         `json:"imei"`
	OID              sql.NullInt64  `json:"oid,omitempty"`
	Name             sql.NullString `json:"name,omitempty"`
	ProviderID       int32          `json:"provider_id"`
	ModerationStatus string         `json:"moderation_status"`
}

func (v Vehicle) MarshalJSON() ([]byte, error) {
	type Alias Vehicle
	aux := &struct {
		OID  interface{} `json:"oid,omitempty"`
		Name interface{} `json:"name,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(&v),
	}

	if !v.OID.Valid {
		aux.OID = nil
	} else {
		aux.OID = v.OID.Int64
	}

	if !v.Name.Valid {
		aux.Name = nil
	} else {
		aux.Name = v.Name.String
	}

	return json.Marshal(aux)
}
