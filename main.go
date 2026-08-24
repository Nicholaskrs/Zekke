package main

import (
	"context"
	"log"
	"os"
	"template-go/core"
	"template-go/core/database"
	"template-go/core/telemetry/logger"
	"template-go/core/telemetry/metric"
	"template-go/core/telemetry/trace"
	routes "template-go/server/router"
	"template-go/util/config"
	"template-go/util/logtrace"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func main() {
	loadConfig, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := logger.NewZerologLogger("Main")

	// Handle CLI subcommands: go run main.go -- migrate | seed | migrate seed
	if args := os.Args[1:]; len(args) > 0 {
		runCommands(args, loadConfig, logger)
		return
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

	router := gin.New()
	db := database.NewEngine(&loadConfig).SetupModels(logger, &loadConfig)
	router.Use(trace.GinHandler)
	router.Use(logger.RouterLogger())
	router.Use(metric.GinMiddleware())

	// TODO: Before deploying to production, set AllowAllOrigins to false
	// and configure HOST_DOMAIN with a comma-separated list of allowed
	// frontend origins, e.g. "https://example.com,https://admin.example.com".
	// AllowAllOrigins + AllowCredentials allows credentialed requests from any origin.
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		//AllowOrigins:     strings.Split(loadConfig.HostDomain, ","),
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

// runCommands handles one-off DB operations (migrate/seed) without booting
// the full server, tracing, or metrics stack.
func runCommands(args []string, cfg config.Config, logger logger.Logger) {
	engine := database.NewEngine(&cfg)
	db := engine.SetupModels(logger, &cfg)

	trace := &logtrace.LogTrace{
		TraceId: uuid.New().String(),
	}

	for _, cmd := range args {
		switch cmd {
		case "migrate":
			if err := engine.MigrateModels(logger, db); err != nil {
				logger.ErrorErr(trace, err).Msg("migration failed")
			}
		case "seed":
			if err := engine.SeedModels(logger); err != nil {
				logger.ErrorErr(trace, err).Msg("seeding failed")
			}
		default:
			logger.Error(trace).Msg("unknown command")
			os.Exit(1)
		}
	}
}
