package metric

import (
	"net/http"
	"template-go/server/middleware"
	"template-go/util/config"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, handler http.Handler, config config.Config) {
	if handler == nil {
		return
	}
	prometheus := router.Group("/prometheus")
	prometheus.Use(middleware.InternalMiddleware(config))
	prometheus.GET("/metrics", gin.WrapH(handler))
}
