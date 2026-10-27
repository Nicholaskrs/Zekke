package service_test

import (
	"testing"

	healthSvc "template-go/modules/health/svc"
)

type healthCheckTestCase struct {
	name        string
	input       *healthSvc.TestHealthIn
	wantSuccess bool
	wantCaller  string
}

func TestHealthCheckService_TestHealth(t *testing.T) {
	t.Parallel()

	tests := []healthCheckTestCase{
		{
			name:        "returns healthy service status",
			input:       &healthSvc.TestHealthIn{Trace: testTrace()},
			wantSuccess: true,
			wantCaller:  "HealthCaller",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := healthSvc.NewHealthCheckService().TestHealth(tt.input)
			assertEqual(t, "TestHealth() success", tt.wantSuccess, got.Success)
			if got.Health == nil {
				t.Fatal("TestHealth() returned nil health")
			}
			assertEqual(t, "TestHealth() caller", tt.wantCaller, got.Health.ServiceCaller)
		})
	}
}
