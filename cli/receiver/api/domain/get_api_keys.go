package domain

import output "github.com/daniil11ru/egts/cli/receiver/api/dto/db/out"

type ApiKeysRepository interface {
	GetApiKeys() ([]output.ApiKey, error)
}

type GetApiKeys struct {
	ApiKeysRepository ApiKeysRepository
}

func (domain *GetApiKeys) Run() ([]output.ApiKey, error) {
	return domain.ApiKeysRepository.GetApiKeys()
}
