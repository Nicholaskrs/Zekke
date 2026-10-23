package user

import (
	userSvc "template-go/modules/user/svc"

	"github.com/gin-gonic/gin"
)

func SetRouterAuthorize(masterGroup *gin.RouterGroup, userService userSvc.UserService) *gin.RouterGroup {
	handler := NewUserHandler(userService)
	group := masterGroup.Group("/auth")
	{
		group.POST("/login", handler.Login)
	}

	return group
}

func SetRouterInternal(masterGroup *gin.RouterGroup, userService userSvc.UserService) *gin.RouterGroup {
	handler := NewUserHandler(userService)
	group := masterGroup.Group("/user")
	{
		group.POST("/register", handler.Register)
	}

	return group
}

func SetRouterAuthenticated(masterGroup *gin.RouterGroup, userService userSvc.UserService) *gin.RouterGroup {
	handler := NewUserHandler(userService)
	group := masterGroup.Group("/user")
	{
		group.POST("/refresh", handler.Refresh)
		group.GET("/profile", handler.GetUserProfile)
		group.POST("/fcm-token/insert", handler.InsertFcmToken)
		group.POST("/fcm-token/delete", handler.DeleteFcmToken)

	}

	return group
}

func SetRouterAuthorizeAdmin(masterGroup *gin.RouterGroup, userAdminService userSvc.UserAdminService) *gin.RouterGroup {
	handler := NewUserAdminHandler(userAdminService)
	group := masterGroup.Group("/auth")
	{
		group.POST("/login", handler.Login)
	}

	return group
}

func SetRouterAuthenticatedAdmin(masterGroup *gin.RouterGroup, userAdminService userSvc.UserAdminService) *gin.RouterGroup {
	handler := NewUserAdminHandler(userAdminService)
	group := masterGroup.Group("/user")
	{

		group.POST("/filter", handler.FilterUser)
		group.DELETE("/user", handler.DeleteUser)
		group.POST("/refresh", handler.Refresh)
		group.POST("/change-password", handler.ChangePassword)
		group.POST("/register", handler.RegisterNewAdmin)

	}

	return group

}
