package repository

import (
	output "github.com/daniil11ru/egts/cli/receiver/api/dto/db/out"
	"github.com/daniil11ru/egts/cli/receiver/api/source"
)

type AdditionalData interface {
	GetApiKeys() ([]output.ApiKey, error)
}

type AdditionalDataSimple struct {
	PostgreSource *source.Postgre
}

func NewAdditionalDataSimple(postgreSource *source.Postgre) *AdditionalDataSimple {
	return &AdditionalDataSimple{PostgreSource: postgreSource}
}

func (r *AdditionalDataSimple) GetApiKeys() ([]output.ApiKey, error) {
	return r.PostgreSource.GetApiKeys()
}
