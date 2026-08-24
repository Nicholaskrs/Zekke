package svc

import (
	"template-go/data/model"
	"template-go/util/logtrace"
)

type HealthCheckService interface {
	TestHealth(*TestHealthIn) *TestHealthOut
}

type TestHealthIn struct {
	Trace *logtrace.LogTrace
}

type TestHealthOut struct {
	Success      bool
	ErrorMessage string
	Health       *model.Health
	ErrorCode    int
}
