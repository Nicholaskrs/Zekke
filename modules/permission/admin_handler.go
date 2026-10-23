package permission

import (
	"net/http"
	"template-go/api"
	adminApi "template-go/api/admin"
	"template-go/base/constants"
	permissionSvc "template-go/modules/permission/svc"
	"template-go/util/logtrace"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PermissionAdminHandler struct {
	PermissionAdminService permissionSvc.PermissionAdminService
}

func NewPermissionHandler(
	PermissionAdminService permissionSvc.PermissionAdminService,
) *PermissionAdminHandler {
	return &PermissionAdminHandler{
		PermissionAdminService: PermissionAdminService,
	}
}

func (h *PermissionAdminHandler) CreatePrivilegeGroup(c *gin.Context) {
	response := &adminApi.CreatePrivilegeGroupResp{}
	trace := logtrace.GetLogTrace(c)
	request := adminApi.CreatePrivilegeGroupReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "privilege group", constants.MethodCreate, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "privilege group", constants.MethodCreate, &request, ve, trace.TraceId)
		return
	}

	// Get current user login id.
	userID := c.GetInt(constants.ContextAdminID)
	privilegeGroup := c.GetString(constants.ContextAdminPrivilegeGroup)

	// Call service
	privilegeOut := h.PermissionAdminService.InsertNewPrivilegeGroup(c,
		&permissionSvc.InsertNewPrivilegeGroupIn{
			Trace:               trace,
			UserAdminId:         userID,
			AdminPrivilegeGroup: privilegeGroup,
			PrivilegeGroup:      request.PrivilegeGroup,
			Description:         request.Description,
		})
	if !privilegeOut.BaseOut.Success {
		api.ReturnResponse(c, "privilege group", constants.MethodCreate, response, privilegeOut.BaseOut.ErrorMessage, privilegeOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	// Send response
	api.ReturnResponse(c, "privilege group", constants.MethodCreate, response, "", http.StatusOK, trace.TraceId)

}

func (h *PermissionAdminHandler) AssignPrivilegeToGroup(c *gin.Context) {
	response := &adminApi.AssignPrivilegeToGroupResp{}
	trace := logtrace.GetLogTrace(c)
	request := adminApi.AssignPrivilegeToGroupReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "assign privilege to privilege group", constants.MethodNull, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "assign privilege to privilege group", constants.MethodNull, &request, ve, trace.TraceId)
		return
	}

	// Get current user login id.
	userID := c.GetInt(constants.ContextAdminID)
	privilegeGroup := c.GetString(constants.ContextAdminPrivilegeGroup)

	// Call service
	privilegeOut := h.PermissionAdminService.AssignPrivilegeToGroup(c,
		&permissionSvc.AssignPrivilegeToGroupIn{
			Trace:                trace,
			UserAdminId:          userID,
			AdminPrivilegeGroup:  privilegeGroup,
			TargetPrivilegeGroup: request.PrivilegeGroup,
			PrivilegeCode:        request.PrivilegeCode,
		})
	if !privilegeOut.BaseOut.Success {
		api.ReturnResponse(c, "assign privilege to privilege group", constants.MethodNull, response, privilegeOut.BaseOut.ErrorMessage, privilegeOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	// Send response
	api.ReturnResponse(c, "assign privilege to privilege group", constants.MethodNull, response, "", http.StatusOK, trace.TraceId)

}

func (h *PermissionAdminHandler) FilterAdminGroup(c *gin.Context) {
	response := &adminApi.FilterAdminGroupResp{}
	trace := logtrace.GetLogTrace(c)
	request := adminApi.FilterAdminGroupReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "privilege group", constants.MethodGet, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "privilege group", constants.MethodGet, &request, ve, trace.TraceId)
		return
	}

	// Get current user login id.
	userID := c.GetInt(constants.ContextAdminID)
	privilegeGroup := c.GetString(constants.ContextAdminPrivilegeGroup)

	// Call service
	privilegeOut := h.PermissionAdminService.FilterPrivilegeGroup(c,
		&permissionSvc.FilterPrivilegeGroupIn{
			Trace:               trace,
			AdminUserId:         userID,
			AdminPrivilegeGroup: privilegeGroup,
			Page:                request.Page,
			Limit:               request.GetLimit(),
		})
	if !privilegeOut.BaseOut.Success {
		api.ReturnResponse(c, "privilege group", constants.MethodGet, response, privilegeOut.BaseOut.ErrorMessage, privilegeOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	response.PrivilegeGroups = adminApi.ParsePrivilegeGroups(privilegeOut.List)
	response.TotalRows = privilegeOut.TotalRows
	response.Limit = request.GetLimit()
	response.Page = request.Page

	// Send response
	api.ReturnResponse(c, "privilege group", constants.MethodGet, response, "", http.StatusOK, trace.TraceId)

}

func (h *PermissionAdminHandler) AssignAdminUserToPrivilegeGroup(c *gin.Context) {
	response := &adminApi.AssignAdminUserToPrivilegeGroupResp{}
	trace := logtrace.GetLogTrace(c)
	request := adminApi.AssignAdminUserToPrivilegeGroupReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "assign privilege to user", constants.MethodNull, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "assign privilege to user", constants.MethodFetch, &request, ve, trace.TraceId)
		return
	}

	// Get current user login id.
	userID := c.GetInt(constants.ContextAdminID)
	privilegeGroup := c.GetString(constants.ContextAdminPrivilegeGroup)

	// Call service
	privilegeOut := h.PermissionAdminService.ChangeAdminPrivilegeRole(c,
		&permissionSvc.ChangeAdminPrivilegeRoleIn{
			Trace:               trace,
			UserAdminId:         userID,
			AdminPrivilegeGroup: privilegeGroup,
			TargetUserID:        request.TargetUserID,
			NewPrivilegeGroup:   request.NewPrivilegeGroup,
		})
	if !privilegeOut.BaseOut.Success {
		api.ReturnResponse(c, "assign privilege to privilege group", constants.MethodNull, response, privilegeOut.BaseOut.ErrorMessage, privilegeOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	// Send response
	api.ReturnResponse(c, "assign privilege to privilege group", constants.MethodNull, response, "", http.StatusOK, trace.TraceId)

}

func (h *PermissionAdminHandler) FilterPrivilegeGroupPrivileges(c *gin.Context) {
	response := &adminApi.FilterPrivilegeGroupPrivilegesResp{}
	trace := logtrace.GetLogTrace(c)
	request := adminApi.FilterPrivilegeGroupPrivilegesReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "privilege", constants.MethodGet, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "privilege", constants.MethodGet, &request, ve, trace.TraceId)
		return
	}

	// Get current user login id.
	userID := c.GetInt(constants.ContextAdminID)
	privilegeGroup := c.GetString(constants.ContextAdminPrivilegeGroup)

	// Call service
	privilegeOut := h.PermissionAdminService.FilterPrivilegeGroupPrivileges(c,
		&permissionSvc.FilterPrivilegeGroupPrivilegesIn{
			Trace:               trace,
			AdminUserId:         userID,
			AdminPrivilegeGroup: privilegeGroup,
			Page:                request.Page,
			Limit:               request.GetLimit(),
			PrivilegeGroup:      request.PrivilegeGroup,
		})
	if !privilegeOut.BaseOut.Success {
		api.ReturnResponse(c, "privilege group", constants.MethodGet, response, privilegeOut.BaseOut.ErrorMessage, privilegeOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	response.Privileges = adminApi.ParsePrivilegeGroupPrivileges(privilegeOut.List)
	response.TotalRows = privilegeOut.TotalRows
	response.Limit = request.GetLimit()
	response.Page = request.Page

	// Send response
	api.ReturnResponse(c, "privilege group", constants.MethodGet, response, "", http.StatusOK, trace.TraceId)

}
