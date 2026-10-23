package core

import (
	healthSvc "template-go/modules/health/svc"
	permissionRepo "template-go/modules/permission/repository"
	permissionSvc "template-go/modules/permission/svc"
	userRepo "template-go/modules/user/repository"
	userSvc "template-go/modules/user/svc"
	"template-go/util/config"
	"template-go/util/storage"

	"gorm.io/gorm"
)

// TODO: If have a new repo or service, add here...
type HandlerDependencies struct {
	Config config.Config

	// Services
	HealthCheckService     healthSvc.HealthCheckService
	UserService            userSvc.UserService
	AwsStorage             storage.Storage
	PermissionAdminService permissionSvc.PermissionAdminService
	UserAdminService       userSvc.UserAdminService
}

func InitClient(config config.Config, DB *gorm.DB) *HandlerDependencies {

	awsStorage := storage.NewAwsStorage(
		config.StorageRegion,
		config.StorageBucket,
		config.StorageBasePath,
		config.StorageBaseUrl,
	)

	// Load Repositories.
	userStorage := userRepo.NewUserStorage(DB)
	permissionStorage := permissionRepo.NewPermissionStorage(DB)

	// Load Services.
	healthCheckService := healthSvc.NewHealthCheckService()
	permissionAdminService := permissionSvc.NewPermissionAdminService(*permissionStorage)
	userService := userSvc.NewUserService(*userStorage, config)
	adminService := userSvc.NewUserAdminService(*userStorage, permissionAdminService, config)

	return &HandlerDependencies{
		Config:                 config,
		HealthCheckService:     healthCheckService,
		UserService:            userService,
		AwsStorage:             awsStorage,
		PermissionAdminService: permissionAdminService,
		UserAdminService:       adminService,
	}
}
