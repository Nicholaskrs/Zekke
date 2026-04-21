package api

import (
	"template-go/data/model"
)

type Health struct {
	ServiceCaller string `json:"service_caller"`
}

func ParseHealth(param *model.Health) *Health {
	return &Health{
		ServiceCaller: param.ServiceCaller,
	}
}
