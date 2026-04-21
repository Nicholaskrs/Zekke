package api

type Auth struct {
	Token    string `json:"token"`
	UserID   uint   `json:"user_id"`
	UserRole string `json:"user_role"`
	FullName string `json:"full_name"`
	Username string `json:"username"`
}
