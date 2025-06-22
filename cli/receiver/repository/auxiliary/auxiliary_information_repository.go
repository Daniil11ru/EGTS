package auxiliary

import (
	aux "github.com/daniil11ru/egts/cli/receiver/source/auxiliary"
)

type AuxiliaryInformationRepository struct {
	Source aux.AuxiliaryInformationSource
}

func (auxInfoRepo *AuxiliaryInformationRepository) GetAllVehicles() ([]aux.Vehicle, error) {
	return auxInfoRepo.Source.GetAllVehicles()
}

func (auxInfoRepo *AuxiliaryInformationRepository) GetAllDirectories() ([]aux.VehicleDirectory, error) {
	return auxInfoRepo.Source.GetAllDirectories()
}

func (auxInfoRepo *AuxiliaryInformationRepository) GetVehiclesByProviderIP(ip string) ([]aux.Vehicle, error) {
	return auxInfoRepo.Source.GetVehiclesByProviderIP(ip)
}

func (auxInfoRepo *AuxiliaryInformationRepository) GetVehicleByOID(OID int32) (aux.Vehicle, error) {
	return auxInfoRepo.Source.GetVehicleByOID(OID)
}
