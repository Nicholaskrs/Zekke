package api

type HealthCheckReq struct {
}

type HealthCheckResp struct {
	*ApiResponse
	Health *Health `json:"health"`
}
