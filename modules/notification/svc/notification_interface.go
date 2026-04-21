package svc

import (
	"context"
	"template-go/data/model"
	"template-go/util/trace"
)

type NotificationService interface {
	SendPushNotification(ctx context.Context, in *SendPushNotificationIn) *SendPushNotificationOut
}

type SendPushNotificationIn struct {
	Trace  *trace.Trace
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
