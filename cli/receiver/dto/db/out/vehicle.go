package out

import (
	"encoding/json"

	"github.com/daniil11ru/egts/cli/receiver/dto/other"
)

type Vehicle struct {
	ID               int32                  `json:"id" gorm:"column:id"`
	IMEI             string                 `json:"imei" gorm:"column:imei"`
	OID              *int64                 `json:"oid,omitempty" gorm:"column:oid"`
	Name             *string                `json:"name,omitempty"`
	ProviderId       int32                  `json:"provider_id"`
	ModerationStatus other.ModerationStatus `json:"moderation_status"`
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

	return json.Marshal(aux)
}
