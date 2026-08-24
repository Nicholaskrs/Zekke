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

// SetupModels only handles the connection + pool config.
func (*MysqlDatabase) SetupModels(log customLog.Logger, config *config.Config) *gorm.DB {
	var err error
	var DB *gorm.DB
	trace := &logtrace.LogTrace{
		TraceId: uuid.New().String(),
	}

	// MySQL DSN format — NOT the same syntax as Postgres.
	connName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=%s",
		config.DbUser,
		config.DbPassword,
		config.DbHost,
		config.DbPort,
		config.DbName,
		config.DbTimezone, // e.g. "Local" or "Asia%2FJakarta" if it contains a slash
	)

	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
	}
	if config.LogType == "1" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	DB, err = gorm.Open(mysql.Open(connName), gormConfig)
	if err != nil {
		log.PanicNoTrace().Msg("Failed to connect to database!")
		panic("Failed to connect to database!")
	}

	if err := DB.Use(tracing.NewPlugin()); err != nil {
		log.WarnNoTrace().Msg(fmt.Sprintf("Failed to attach telemetry plugin: %v", err))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.FatalErr(trace, err).Msg("failed to get SQL DB")
		panic("Get SQL DB failed")
	}

	sqlDB.SetMaxOpenConns(config.DbMaxOpenConns)
	sqlDB.SetMaxIdleConns(config.DbMaxIdleConns)
	sqlDB.SetConnMaxLifetime(15 * time.Minute)

	return DB
}

// MigrateModels runs AutoMigrate. Call this explicitly via the migrate command.
func (*MysqlDatabase) MigrateModels(log customLog.Logger, DB *gorm.DB) error {
	trace := &logtrace.LogTrace{TraceId: uuid.New().String()}

	err := DB.AutoMigrate(
		// @Notes: Add model in here
		&model.FcmToken{},
		&model.User{},
	)
	if err != nil {
		log.ErrorErr(trace, err).Msg("Migration Failed")
		return err
	}
	return nil
}

// SeedModels runs the seeder. Call this explicitly via the seed command.
func (*MysqlDatabase) SeedModels(log customLog.Logger) error {
	trace := &logtrace.LogTrace{TraceId: uuid.New().String()}

	if err := initSeeder(); err != nil {
		log.ErrorErr(trace, err).Msg("init Seeder failed")
		return err
	}
	return nil
}
