package primary

import (
	connector "github.com/daniil11ru/egts/cli/receiver/connector"
	"github.com/daniil11ru/egts/cli/receiver/repository/primary/types"
)

type PrimarySource interface {
	Initialize(connector connector.Connector)
	GetAllVehicles() ([]types.Vehicle, error)
	GetAllProviders() ([]types.Provider, error)
	GetVehicleByID(id int32) (types.Vehicle, error)
	GetVehiclesByProviderIP(ip string) ([]types.Vehicle, error)
	GetVehicleByOID(OID int32) (types.Vehicle, error)
	GetVehicleByOIDAndProviderID(OID int32, providerID int32) (types.Vehicle, error)
	AddVehicle(v types.Vehicle) (int32, error)
	UpdateVehicleOID(id int32, OID int32) error
	GetProviderByIP(ip string) (types.Provider, error)
	GetAllIPs() ([]string, error)
	AddVehicleMovement(message interface{ ToBytes() ([]byte, error) }, vehicleID int) (int32, error)
	GetLastVehiclePosition(vehicle_id int32) (types.Position, error)
}
