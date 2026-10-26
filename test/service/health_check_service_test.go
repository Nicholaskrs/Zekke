package service_test

import (
	"testing"

	healthSvc "template-go/modules/health/svc"
)

func TestHealthCheckService_TestHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       *healthSvc.TestHealthIn
		wantSuccess bool
		wantCaller  string
	}{
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
			if got.Success != tt.wantSuccess {
				t.Fatalf("TestHealth() success = %t, want %t", got.Success, tt.wantSuccess)
			}
			if got.Health == nil || got.Health.ServiceCaller != tt.wantCaller {
				t.Fatalf("TestHealth() health = %+v, want caller %q", got.Health, tt.wantCaller)
			}
		})
	}
}
