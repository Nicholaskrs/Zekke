package routes

import (
	"template-go/core"
	"template-go/modules/health"
	"template-go/modules/user"
	"template-go/server/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, deps *core.HandlerDependencies) {
	router.RedirectTrailingSlash = true

	// @Notes: inject dependency

	// Public routes
	r := router.Group("/api")
	v1 := r.Group("/v1")

	// No Authentication.
	unAuthenticatedRoutes(deps, v1)

	// Require Authentication Routes.
	authenticatedRoutes(deps, v1)

	// Public routes
	rInternal := router.Group("/api-internal")
	v1Interal := rInternal.Group("/v1")

	// Require X-API-Key Routes. For internal uses only
	internalAuthenticatedRoutes(deps, v1Interal)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": 404, "message": "API Not found"})
	})
}

func unAuthenticatedRoutes(deps *core.HandlerDependencies, v1 *gin.RouterGroup) {
	health.SetRouter(v1, deps.HealthCheckService)
	user.SetRouterAuthorize(v1, deps.UserService)

}

// Use jwt authentication
func authenticatedRoutes(deps *core.HandlerDependencies, v1 *gin.RouterGroup) {
	v1.Use(middleware.AuthorizeJWT(deps.Config))
	health.AuthenticatedSetRouter(v1, deps.HealthCheckService)
	user.SetRouterAuthenticated(v1, deps.UserService)
}

// Use X-API-Key authentication
func internalAuthenticatedRoutes(deps *core.HandlerDependencies, v1 *gin.RouterGroup) {
	v1.Use(middleware.InternalMiddleware(deps.Config))
	user.SetRouterAdmin(v1, deps.UserService)
}
