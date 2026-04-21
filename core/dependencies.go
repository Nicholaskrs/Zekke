package core

import (
	healthSvc "template-go/modules/health/svc"
	userRepo "template-go/modules/user/repository"
	userSvc "template-go/modules/user/svc"
	"template-go/util/config"
	"template-go/util/storage"

	"gorm.io/gorm"
)

// TODO: If have a new repo or service, add here...
type HandlerDependencies struct {
	Config config.Config
	DB     *gorm.DB

	// Services
	HealthCheckService healthSvc.HealthCheckService
	UserService        userSvc.UserService
	AwsStorage         storage.Storage
}

func InitClient(config config.Config, DB *gorm.DB) *HandlerDependencies {

	awsStorage := storage.NewAwsStorage(
		config.StorageRegion,
		config.StorageBucket,
		config.StorageBasePath,
		config.StorageBaseUrl,
	)

	// Load Repositories
	userStorage := userRepo.NewUserStorage(DB)

	// Load Services
	healthCheckService := healthSvc.NewHealthCheckService()
	userService := userSvc.NewUserService(*userStorage, config)

	return &HandlerDependencies{
		Config:             config,
		HealthCheckService: healthCheckService,
		UserService:        userService,
		AwsStorage:         awsStorage,
	}
}
