package svc

import (
	"context"
	"template-go/data/model"
	"template-go/util/logtrace"
)

type NotificationService interface {
	SendPushNotification(ctx context.Context, in *SendPushNotificationIn) *SendPushNotificationOut
}

type SendPushNotificationIn struct {
	Trace  *logtrace.LogTrace
	UserID uint
	Title  string
	Body   string
	Data   map[string]string
}

type SendPushNotificationOut struct {
	Success      bool
	ErrorMessage string
	Health       *model.Health
	ErrorCode    int
}
