package database

import (
	"fmt"
	customLog "template-go/core/telemetry/logger"
	"template-go/data/model"
	"template-go/util/config"
	"template-go/util/logtrace"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

var _ Database = (*PsqlDatabase)(nil)

type PsqlDatabase struct{}

func (*PsqlDatabase) SetupModels(log customLog.Logger, config *config.Config) *gorm.DB {
	var err error
	var DB *gorm.DB
	trace := &logtrace.LogTrace{
		TraceId: uuid.New().String(),
	}

	connName := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		config.DbHost,
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbPort,
		"Asia/Jakarta",
	)

	gormConfig := &gorm.Config{}
	if config.LogType == "1" {
		gormConfig = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info), // Enable query logging
		}
	}

	DB, err = gorm.Open(postgres.Open(connName), gormConfig)
	if err != nil {
		log.PanicNoTrace().Msg("Failed to connect to database!")
		panic("Failed to connect to database!")
	}

	if err := DB.Use(tracing.NewPlugin()); err != nil {
		log.WarnNoTrace().Msg(fmt.Sprintf("Failed to attach telemetry plugin: %v", err))
	}

	err = DB.AutoMigrate(
		// @Notes: Add model in here
		&model.AuditLog{},
		&model.FcmToken{},
		&model.User{},
	)

	if err != nil {
		log.ErrorErr(trace, err).Msg("Migration Failed")
		panic("Migration Failed!")
	}

	// Get the raw SQL DB object for connection pooling
	sqlDB, err := DB.DB()
	if err != nil {
		log.FatalErr(trace, err).Msg("failed to get PSQL DB")
		panic("Get SQL DB failed")

	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(10)               // Max open connections
	sqlDB.SetMaxIdleConns(5)                // Max idle connections
	sqlDB.SetConnMaxLifetime(1 * time.Hour) // Max connection lifetime

	// Call Seeder
	err = initSeeder()
	if err != nil {
		log.ErrorErr(trace, err).Msg("init Seeder failed")
		panic("init Seeder failed")
	}

	return DB
}
