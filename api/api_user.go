package api

type UserRegisterReq struct {
	Username string `json:"username" binding:"required,ascii"`
	Password string `json:"password" binding:"required,ascii"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required,ascii"`
	UserRole string `json:"user_role" binding:"required,ascii"`
}

type UserRegisterResp struct {
	*ApiResponse
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
