package domain

import (
	"github.com/daniil11ru/egts/cli/receiver/repository/primary"
)

type GetIPWhiteList struct {
	PrimaryRepository primary.PrimaryRepository
}

func (domain *GetIPWhiteList) Run() ([]string, error) {
	return domain.PrimaryRepository.Source.GetAllIPs()
}
