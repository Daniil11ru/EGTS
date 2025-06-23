package auxiliary

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	connector "github.com/daniil11ru/egts/cli/receiver/connector"
)

type ModerationStatus string

const (
	ModerationStatusPending  ModerationStatus = "pending"
	ModerationStatusApproved ModerationStatus = "approved"
	ModerationStatusRejected ModerationStatus = "rejected"
)

func (ms *ModerationStatus) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan ModerationStatus from %T", value)
	}
	*ms = ModerationStatus(string(b))
	return nil
}

func (ms ModerationStatus) Value() (driver.Value, error) {
	return string(ms), nil
}

type Vehicle struct {
	ID                 int32
	IMEI               int64
	OID                sql.NullInt32
	LicensePlateNumber sql.NullString
	ProviderID         int32
	ModerationStatus   ModerationStatus
}

type Provider struct {
	ID   int32
	Name string
	IP   []string
}

type AuxiliaryInformationSource interface {
	Initialize(connector connector.Connector)
	GetAllVehicles() ([]Vehicle, error)
	GetAllProviders() ([]Provider, error)
	GetVehicleByID(id int32) (Vehicle, error)
	GetVehiclesByProviderIP(ip string) ([]Vehicle, error)
	GetVehicleByOID(OID int32) (Vehicle, error)
	GetVehicleByOIDAndProviderID(OID int32, providerID int32) (Vehicle, error)
	AddVehicle(v Vehicle) (int32, error)
	UpdateVehicleOID(id int32, OID int32) error
	GetProviderByIP(ip string) (Provider, error)
	GetAllIPs() ([]string, error)
}
