package svc

import (
	"template-go/data/model"
	"template-go/util/trace"
)

type HealthCheckService interface {
	TestHealth(*TestHealthIn) *TestHealthOut
}

type TestHealthIn struct {
	Trace *trace.Trace
}

type TestHealthOut struct {
	Success      bool
	ErrorMessage string
	Health       *model.Health
	ErrorCode    int
}
