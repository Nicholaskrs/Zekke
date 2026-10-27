package mocks

//go:generate go run github.com/matryer/moq@v0.8.1 -pkg mocks -out user_service.go ../../modules/user/svc UserService
//go:generate go run github.com/matryer/moq@v0.8.1 -pkg mocks -out permission_admin_service.go ../../modules/permission/svc PermissionAdminService
//go:generate go run github.com/matryer/moq@v0.8.1 -pkg mocks -out firebase_sender.go ../../util/firebase Sender
