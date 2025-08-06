package repository

import (
	"github.com/daniil11ru/egts/cli/receiver/api/model"
	"github.com/daniil11ru/egts/cli/receiver/api/source"
)

type AdditionalData interface {
	GetApiKeys() ([]model.ApiKey, error)
}

type AdditionalDataSimple struct {
	PostgreSource *source.Postgre
}

func NewAdditionalDataSimple(postgreSource *source.Postgre) *AdditionalDataSimple {
	return &AdditionalDataSimple{PostgreSource: postgreSource}
}

func (r *AdditionalDataSimple) GetApiKeys() ([]model.ApiKey, error) {
	return r.PostgreSource.GetApiKeys()
}
