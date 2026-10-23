package api

type AuthReq struct {
	Username string `json:"username" binding:"required,ascii"`
	Password string `json:"password" binding:"required,ascii"`
}

type AuthResp struct {
	*Auth
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResp struct {
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ChangePasswordResp struct {
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type RegisterNewAdminReq struct {
	Username       string `json:"username" binding:"required,ascii"`
	Password       string `json:"password" binding:"required,ascii"`
	Email          string `json:"email" binding:"required,email"`
	FullName       string `json:"full_name" binding:"required,ascii"`
	PrivilegeGroup string `json:"privilege_group" binding:"required,ascii"`
}

type RegisterNewAdminResp struct {
}
