package health

import (
	"net/http"
	"template-go/api"
	"template-go/base/constants"
	"template-go/modules/health/svc"
	"template-go/util/trace"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type HealthCheckHandler struct {
	HealthCheckService svc.HealthCheckService
}

func NewHealthCheckHandler(
	HealthCheckService svc.HealthCheckService,
) *HealthCheckHandler {
	return &HealthCheckHandler{
		HealthCheckService: HealthCheckService,
	}
}

// TestHealth used to check service by return success.
func (handler *HealthCheckHandler) TestHealth(c *gin.Context) {
	response := &api.HealthCheckResp{}
	trace := &trace.Trace{
		TraceId: uuid.New().String(),
	}

	request := api.HealthCheckReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "health", constants.MethodGet, request, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "health", constants.MethodGet, &request, ve, trace.TraceId)
		return
	}

	// Call service
	healthCheckOut := handler.HealthCheckService.TestHealth(&svc.TestHealthIn{Trace: trace})
	if !healthCheckOut.Success {
		api.ReturnResponse(c, "health", constants.MethodGet, response, healthCheckOut.ErrorMessage, healthCheckOut.ErrorCode, trace.TraceId)
		return
	}

	health := api.ParseHealth(healthCheckOut.Health)
	response.Health = health

	// Send response
	api.ReturnResponse(c, "health", constants.MethodGet, response, "", http.StatusOK, trace.TraceId)
}
