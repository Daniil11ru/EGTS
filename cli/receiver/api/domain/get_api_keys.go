package domain

import "github.com/daniil11ru/egts/cli/receiver/api/model"

type ApiKeysRepository interface {
	GetApiKeys() ([]model.ApiKey, error)
}

type GetApiKeys struct {
	ApiKeysRepository ApiKeysRepository
}

func (domain *GetApiKeys) Run() ([]model.ApiKey, error) {
	return domain.ApiKeysRepository.GetApiKeys()
}
