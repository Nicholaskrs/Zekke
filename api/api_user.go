package api

type UserRegisterReq struct {
	Username      string `json:"username" binding:"required,ascii"`
	Password      string `json:"password" binding:"required,ascii"`
	Email         string `json:"email" binding:"required,email"`
	FullName      string `json:"full_name" binding:"required,ascii"`
	UserRole      string `json:"user_role" binding:"required,ascii"`
	DistributorID uint   `json:"distributor_id" binding:"required,min=1"`
	AreaID        uint   `json:"area_id" binding:"required,min=1"`
}

type UserRegisterResp struct {
	*ApiResponse
}

type GetUserSalesReq struct {
}

type GetUserResp struct {
	*ApiResponse
	Users []*User `json:"user"`
}

type ChangePasswordReq struct {
	ExternalID string `json:"external_id" binding:"required,ascii"`
	Password   string `json:"password" binding:"required,ascii"`
}
type GetUserProfileResp struct {
	*ApiResponse
	User *User `json:"user"`
}

type InsertFcmTokenReq struct {
	Token string `json:"token" binding:"required"`
}
type InsertFcmTokenResp struct {
	*ApiResponse
}

type DeleteFcmTokenReq struct {
	Token string `json:"token" binding:"required"`
}
type DeleteFcmTokenResp struct {
	*ApiResponse
}
