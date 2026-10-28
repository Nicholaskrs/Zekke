package service_test

import (
	"testing"

	healthSvc "template-go/modules/health/svc"
)

func TestHealthCheckService_TestHealth(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		caller  string
	}

	type testCase struct {
		name     string
		input    *healthSvc.TestHealthIn
		expected expected
	}

	// Test cases:
	//
	// +----------------------------+---------+---------------+
	// | Scenario                   | Success | Caller        |
	// +----------------------------+---------+---------------+
	// | default health check input | true    | HealthCaller  |
	// +----------------------------+---------+---------------+
	tests := []testCase{
		{
			name:     "returns healthy service status",
			input:    &healthSvc.TestHealthIn{Trace: testTrace()},
			expected: expected{success: true, caller: "HealthCaller"},
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
