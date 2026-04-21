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

func SetRouterAdmin(masterGroup *gin.RouterGroup, userService userSvc.UserService) *gin.RouterGroup {
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
		group.POST("/change-password", handler.ChangePasswordByExternalID)
		group.GET("/sales/profile", handler.GetUserProfile)
		group.POST("/fcm-token/insert", handler.InsertFcmToken)
		group.POST("/fcm-token/delete", handler.DeleteFcmToken)

	}

	return group
}
