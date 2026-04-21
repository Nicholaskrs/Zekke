package svc

import (
	"context"
	"firebase.google.com/go/v4/messaging"
	"net/http"
	userSvc "template-go/modules/user/svc"
	"template-go/util/firebase"
	"template-go/util/logger"
)

type NotificationServiceImpl struct {
	firebaseClient *firebase.Client
	userService    userSvc.UserService
	logger         logger.Logger
}

func NewNotificationService(
	firebaseClient *firebase.Client,
	userService userSvc.UserService,
) NotificationService {
	return &NotificationServiceImpl{
		firebaseClient: firebaseClient,
		userService:    userService,
		logger:         logger.NewZerologLogger("NotificationService"),
	}
}

func (service *NotificationServiceImpl) SendPushNotification(ctx context.Context, in *SendPushNotificationIn) *SendPushNotificationOut {
	resp := &SendPushNotificationOut{}
	service.logger.Info(in.Trace).Msg("Send Push Notification")
	if in.Title == "" {
		service.logger.Warn(in.Trace).Msg("Title is empty")
		resp.ErrorMessage = "Title is empty"
		resp.ErrorCode = http.StatusUnprocessableEntity
		return resp
	}
	if in.Body == "" {
		service.logger.Warn(in.Trace).Msg("Body is empty")
		resp.ErrorMessage = "Body is empty"
		resp.ErrorCode = http.StatusUnprocessableEntity
		return resp
	}
	if in.UserID == 0 {
		service.logger.Warn(in.Trace).Msg("UserID is empty")
		resp.ErrorMessage = "UserID is empty"
		resp.ErrorCode = http.StatusUnprocessableEntity
		return resp
	}

	getFcmTokenOut := service.userService.GetUserFcmToken(ctx, &userSvc.GetUserFcmTokenIn{
		Trace:  in.Trace,
		UserID: in.UserID,
	})

	if getFcmTokenOut.Success == false {
		resp.ErrorMessage = getFcmTokenOut.ErrorMessage
		resp.ErrorCode = getFcmTokenOut.ErrorCode
		return resp
	}
	tokens := getFcmTokenOut.FcmTokens
	shouldRemovedTokens := []string{}
	for _, token := range tokens {
		_, err := service.firebaseClient.Send(ctx, in.Title, in.Body, in.Data, token.FcmToken)
		if err != nil {
			if messaging.IsUnregistered(err) {
				shouldRemovedTokens = append(shouldRemovedTokens, token.FcmToken)
			}
		}
	}

	if len(shouldRemovedTokens) != 0 {
		removeFcmTokenOut := service.userService.DeleteFcmTokenBulk(ctx, &userSvc.DeleteFcmTokenBulkIn{
			Trace:  in.Trace,
			Tokens: shouldRemovedTokens,
		})
		if removeFcmTokenOut.Success == false {
			resp.ErrorMessage = removeFcmTokenOut.ErrorMessage
			resp.ErrorCode = removeFcmTokenOut.ErrorCode
			return resp
		}
	}

	resp.Success = true
	return resp
}
