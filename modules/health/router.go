package health

import (
	"template-go/modules/health/svc"

	"github.com/gin-gonic/gin"
)

func SetRouter(masterGroup *gin.RouterGroup, healthService svc.HealthCheckService) *gin.RouterGroup {
	handler := NewHealthCheckHandler(healthService)
	group := masterGroup.Group("/health")
	{
		group.GET("", handler.TestHealth)
	}

	return group
}

func AuthenticatedSetRouter(masterGroup *gin.RouterGroup, healthService svc.HealthCheckService) *gin.RouterGroup {
	handler := NewHealthCheckHandler(healthService)
	group := masterGroup.Group("/auth/health")
	{
		group.GET("", handler.TestHealth)
	}

	return group
}
