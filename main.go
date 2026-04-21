package main

import (
	"log"
	"template-go/core"
	"template-go/core/database"
	routes "template-go/server/router"
	"template-go/util/config"
	"template-go/util/logger"
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

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	logger := logger.NewZerologLogger("Main")

	router := gin.New()
	db := database.SetupModels(logger, loadConfig)
	router.Use(logger.RouterLogger())

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
	router.Run(loadConfig.ServerPort)
}
