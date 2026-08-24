package main

import (
	"context"
	"log"
	"template-go/core"
	"template-go/core/database"
	"template-go/core/telemetry/logger"
	"template-go/core/telemetry/metric"
	"template-go/core/telemetry/trace"
	routes "template-go/server/router"
	"template-go/util/config"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func main() {
	loadConfig, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	ctx := context.Background()
	shutdown, err := trace.Init(ctx, loadConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown(ctx)

	metricsHandler, err := metric.Init(ctx, loadConfig)
	if err != nil {
		log.Fatal(err)
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	logger := logger.NewZerologLogger("Main")

	router := gin.New()
	db := database.NewEngine(&loadConfig).SetupModels(logger, &loadConfig)
	router.Use(trace.GinHandler)
	router.Use(logger.RouterLogger())
	router.Use(metric.GinMiddleware())

	// TODO: Before deploying to production, replace AllowAllOrigins with an explicit
	// AllowOrigins list for your frontend's domain(s). AllowAllOrigins + AllowCredentials
	// together means any website can make credentialed requests to this API.
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	deps := core.InitClient(loadConfig, db)

	routes.RegisterRoutes(router, deps)
	metric.RegisterRoutes(router, metricsHandler, loadConfig)
	router.Run(loadConfig.ServerPort)
}
