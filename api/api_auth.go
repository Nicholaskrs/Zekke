package api

type ApiAuthReq struct {
	Username string `json:"username" binding:"required,ascii"`
	Password string `json:"password" binding:"required,ascii"`
}

type ApiAuthResp struct {
	*ApiResponse
	Auth *Auth `json:"auth"`
}
