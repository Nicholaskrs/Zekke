package database

import (
	"fmt"
	customLog "template-go/core/telemetry/logger"
	"template-go/util/config"

	"gorm.io/gorm"
)

type Database interface {
	SetupModels(log customLog.Logger, config *config.Config) *gorm.DB
	MigrateModels(log customLog.Logger, DB *gorm.DB) error
	SeedModels(log customLog.Logger) error
}

func NewEngine(cfg *config.Config) Database {
	switch cfg.DbDriver {
	case "postgres":
		return &PsqlDatabase{}
	case "mysql":
		return &MysqlDatabase{}
	default:
		panic(fmt.Sprintf("unsupported db driver: %s", cfg.DbDriver))
	}
}
