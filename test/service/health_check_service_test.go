package service_test

import (
	"testing"

	healthSvc "template-go/modules/health/svc"
)

type healthCheckExpected struct {
	success bool
	caller  string
}

type healthCheckTestCase struct {
	name     string
	input    *healthSvc.TestHealthIn
	expected healthCheckExpected
}

func TestHealthCheckService_TestHealth(t *testing.T) {
	t.Parallel()

	tests := []healthCheckTestCase{
		{
			name:     "returns healthy service status",
			input:    &healthSvc.TestHealthIn{Trace: testTrace()},
			expected: healthCheckExpected{success: true, caller: "HealthCaller"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := healthSvc.NewHealthCheckService().TestHealth(tt.input)
			assertEqual(t, "TestHealth() success", tt.expected.success, got.Success)
			if got.Health == nil {
				t.Fatal("TestHealth() returned nil health")
			}
			assertEqual(t, "TestHealth() caller", tt.expected.caller, got.Health.ServiceCaller)
		})
	}
}
