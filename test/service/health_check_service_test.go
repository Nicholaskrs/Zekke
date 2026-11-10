package service_test

import (
	"testing"

	healthSvc "template-go/modules/health/svc"
	"template-go/test"
	"template-go/test/fixture"
	"template-go/util/logtrace"
)

func TestHealthCheckService_TestHealth(t *testing.T) {
	t.Parallel()

	type input struct {
		Trace *logtrace.LogTrace
	}
	type expected struct {
		success       bool
		serviceCaller string
	}
	type testCase struct {
		name     string
		input    input
		expected expected
	}
	tt := testCase{
		name:     "returns health status",
		input:    input{Trace: fixture.TestTrace()},
		expected: expected{success: true, serviceCaller: "HealthCaller"},
	}

	t.Run(tt.name, func(t *testing.T) {
		got := healthSvc.NewHealthCheckService().TestHealth(&healthSvc.TestHealthIn{Trace: tt.input.Trace})
		test.AssertEqual(t, "TestHealth success", tt.expected.success, got.Success)
		test.AssertTrue(t, "TestHealth health response", got.Health != nil)
		test.AssertEqual(t, "TestHealth service caller", tt.expected.serviceCaller, got.Health.ServiceCaller)
	})
}
