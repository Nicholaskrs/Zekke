package api

import (
	"template-go/api"
)

type HealthCheckReq struct {
}

type HealthCheckResp struct {
	*api.Response
	Health *Health `json:"health"`
}
