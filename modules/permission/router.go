package permission

import (
	permissionSvc "template-go/modules/permission/svc"

	"github.com/gin-gonic/gin"
)

func SetRouterAdmin(masterGroup *gin.RouterGroup, permissionAdminService permissionSvc.PermissionAdminService) *gin.RouterGroup {
	handler := NewPermissionHandler(permissionAdminService)
	group := masterGroup.Group("/permission")
	{
		masterGroup.POST("/privilege-group/filter", handler.FilterAdminGroup)
		masterGroup.POST("/privilege-group/create", handler.CreatePrivilegeGroup)
		masterGroup.POST("/privilege-group/assign", handler.AssignAdminUserToPrivilegeGroup)
		masterGroup.POST("/privilege/assign", handler.AssignPrivilegeToGroup)
		masterGroup.POST("/privilege/filter", handler.FilterPrivilegeGroupPrivileges)
	}
	return group
}
