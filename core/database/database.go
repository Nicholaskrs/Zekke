package database

import (
	"fmt"
	"net/url"
	"template-go/data/model"
	"template-go/util/config"
	customLog "template-go/util/logger"
	trace "template-go/util/trace"
	"time"

	"github.com/google/uuid"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupModels(log customLog.Logger, config config.Config) *gorm.DB {
	var err error
	var DB *gorm.DB
	trace := &trace.Trace{
		TraceId: uuid.New().String(),
	}

	connName := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=%v",
		config.DbUser,
		config.DbPassword,
		config.DbHost,
		config.DbPort,
		config.DbName,
		url.QueryEscape("Asia/Jakarta"),
	)

	log.InfoNoTrace().Msg(fmt.Sprintf("conname is %v", connName))

	gormConfig := &gorm.Config{}
	if config.LogType == "1" {
		gormConfig = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info), // Enable query logging
		}
	}

	DB, err = gorm.Open(mysql.Open(connName), gormConfig)
	if err != nil {
		log.PanicNoTrace().Msg("Failed to connect to database!")
		panic("Failed to connect to database!")
	}
	err = DB.AutoMigrate(
		// @Notes: Add model in here
		&model.Area{},
		&model.Attendance{},
		&model.AuditLog{},
		&model.DistributorProduct{},
		&model.Distributor{},
		&model.FcmToken{},
		&model.Product{},
		&model.PurchaseOrder{},
		&model.PurchaseOrderDetail{},
		&model.PurchaseOrderReturn{},
		&model.Store{},
		&model.User{},
		&model.VisitationImage{},
		&model.Visitation{},
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
