package primary

import (
	"time"

	connector "github.com/daniil11ru/egts/cli/receiver/connector"
	"github.com/daniil11ru/egts/cli/receiver/repository/util"
	"github.com/daniil11ru/egts/cli/receiver/types"
)

type PrimarySource interface {
	Initialize(connector connector.Connector)
	GetAllVehicles() ([]types.Vehicle, error)
	GetAllProviders() ([]types.Provider, error)
	GetVehicleByID(id int32) (types.Vehicle, error)
	GetVehiclesByProviderIP(ip string) ([]types.Vehicle, error)
	GetVehiclesByProviderID(providerID int32) ([]types.Vehicle, error)
	GetVehicleByOID(OID uint32) (types.Vehicle, error)
	GetVehiclesByOIDAndProviderID(OID uint32, providerID int32) ([]types.Vehicle, error)
	AddVehicle(v types.Vehicle) (int32, error)
	UpdateVehicleOID(id int32, OID uint32) error
	GetProviderByIP(ip string) (types.Provider, error)
	GetAllIPs() ([]string, error)
	AddVehicleMovement(data *util.NavigationRecord, vehicleID int) (int32, error)
	GetLastVehiclePosition(vehicleID int32) (types.Position3D, error)
	GetTracks2DOfAllVehicles(after, before time.Time) ([]types.Track2D, error)
	DeleteVehicleMovement(vehicleMovementID int32) error
}
