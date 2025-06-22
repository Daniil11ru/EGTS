package auxiliary

import (
	connector "github.com/daniil11ru/egts/cli/receiver/connector"
)

type Vehicle struct {
	ID                 int
	IMEI               int64
	LicensePlateNumber string
	DirectoryID        int
}

type VehicleDirectory struct {
	ID         int
	ProviderID int
}

type Provider struct {
	ID   int
	Name string
	IP   []string
}

type AuxiliaryInformationSource interface {
	Initialize(connector connector.Connector)
	GetAllVehicles() ([]Vehicle, error)
	GetAllDirectories() ([]VehicleDirectory, error)
	GetAllProviders() ([]Provider, error)
	GetVehiclesByProviderIP(ip string) ([]Vehicle, error)
	GetAllIPs() ([]string, error)
}
