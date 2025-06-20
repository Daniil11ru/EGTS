package domain

import (
	aux "github.com/daniil11ru/egts/cli/receiver/repository/auxiliary"
)

type GetIPWhiteList struct {
	AuxiliaryInformationRepository aux.AuxiliaryInformationRepository
}

func (domain *GetIPWhiteList) Run() ([]string, error) {
	return domain.AuxiliaryInformationRepository.Source.GetAllIPs()
}
