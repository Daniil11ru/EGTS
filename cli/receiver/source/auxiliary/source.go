package auxiliary

import (
	"database/sql"

	connector "github.com/daniil11ru/egts/cli/receiver/connector"
)

type Vehicle struct {
	ID                 int32
	IMEI               int64
	OID                sql.NullInt32
	LicensePlateNumber sql.NullString
	ProviderID         int32
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
	GetVehiclesByProviderIP(ip string) ([]Vehicle, error)
	GetVehicleByOID(OID int32) (Vehicle, error)
	GetVehicleByOIDAndProviderID(OID int32, providerID int32) (Vehicle, error)
	AddVehicle(v Vehicle) (int32, error)
	UpdateVehicleOID(id int32, OID int32) error
	GetProviderByIP(ip string) (Provider, error)
	GetAllIPs() ([]string, error)
}
