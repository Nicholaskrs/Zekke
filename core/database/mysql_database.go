package database

import (
	"fmt"
	customLog "template-go/core/telemetry/logger"
	"template-go/data/model"
	"template-go/util/config"
	"template-go/util/logtrace"
	"time"

	"github.com/google/uuid"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

var _ Database = (*MysqlDatabase)(nil)

type MysqlDatabase struct{}

func (*MysqlDatabase) SetupModels(log customLog.Logger, config *config.Config) *gorm.DB {
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
		config.DbTimezone,
	)

	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
	}
	if config.LogType == "1" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info) // Enable query logging
	}

	DB, err = gorm.Open(mysql.Open(connName), gormConfig)
	if err != nil {
		log.PanicNoTrace().Msg("Failed to connect to database!")
		panic("Failed to connect to database!")
	}

	if err := DB.Use(tracing.NewPlugin()); err != nil {
		log.WarnNoTrace().Msg(fmt.Sprintf("Failed to attach telemetry plugin: %v", err))
	}

	err = DB.AutoMigrate(
		// @Notes: Add model in here
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
		log.FatalErr(trace, err).Msg("failed to get SQL DB")
		panic("Get SQL DB failed")

	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.DbMaxOpenConns) // Max open connections
	sqlDB.SetMaxIdleConns(config.DbMaxIdleConns) // Max idle connections
	sqlDB.SetConnMaxLifetime(15 * time.Minute)   // Max connection lifetime

	// Call Seeder
	err = initSeeder()
	if err != nil {
		log.ErrorErr(trace, err).Msg("init Seeder failed")
		panic("init Seeder failed")
	}

	return DB
}
