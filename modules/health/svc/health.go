package svc

import "template-go/data/model"

type HealthCheckServiceImpl struct {
}

func NewHealthCheckService() HealthCheckService {
	return &HealthCheckServiceImpl{}
}

// TestHealth used to check service by return success
func (service *HealthCheckServiceImpl) TestHealth(*TestHealthIn) *TestHealthOut {
	//logic dst
	return &TestHealthOut{
		Success: true,
		Health: &model.Health{
			ServiceCaller: "HealthCaller",
		},
	}
}
