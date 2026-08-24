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

// SetupModels only handles the connection + pool config.
// It no longer runs migrations or seeding automatically.
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
		config.DbTimezone,
	)

	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
	}
	if config.LogType == "1" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	DB, err = gorm.Open(postgres.Open(connName), gormConfig)
	if err != nil {
		log.PanicNoTrace().Msg("Failed to connect to database!")
		panic("Failed to connect to database!")
	}

	if err := DB.Use(tracing.NewPlugin()); err != nil {
		log.WarnNoTrace().Msg(fmt.Sprintf("Failed to attach telemetry plugin: %v", err))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.FatalErr(trace, err).Msg("failed to get PSQL DB")
		panic("Get SQL DB failed")
	}

	sqlDB.SetMaxOpenConns(config.DbMaxOpenConns)
	sqlDB.SetMaxIdleConns(config.DbMaxIdleConns)
	sqlDB.SetConnMaxLifetime(15 * time.Minute)

	return DB
}

// MigrateModels runs AutoMigrate. Call this explicitly via the migrate command.
func (*PsqlDatabase) MigrateModels(log customLog.Logger, DB *gorm.DB) error {
	trace := &logtrace.LogTrace{TraceId: uuid.New().String()}

	err := DB.AutoMigrate(
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
func (*PsqlDatabase) SeedModels(log customLog.Logger) error {
	trace := &logtrace.LogTrace{TraceId: uuid.New().String()}

	if err := initSeeder(); err != nil {
		log.ErrorErr(trace, err).Msg("init Seeder failed")
		return err
	}
	return nil
}
