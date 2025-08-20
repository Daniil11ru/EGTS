package repository

import (
	output "github.com/daniil11ru/egts/cli/receiver/dto/db/out"
	"github.com/daniil11ru/egts/cli/receiver/source"
)

type AdditionalData interface {
	GetApiKeys() ([]output.ApiKey, error)
}

type AdditionalDataDefault struct {
	Source source.Primary
}

func NewAdditionalDataDefault(source source.Primary) *AdditionalDataDefault {
	return &AdditionalDataDefault{Source: source}
}

func (r *AdditionalDataDefault) GetApiKeys() ([]output.ApiKey, error) {
	return r.Source.GetApiKeys()
}
