package service_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	dto "template-go/base/base"
	"template-go/data/model"
	notificationSvc "template-go/modules/notification/svc"
	userSvc "template-go/modules/user/svc"
	"template-go/test"
	"template-go/test/fixture"
	"template-go/test/mocks"
)

func TestNotificationService_SendPushNotification(t *testing.T) {
	t.Parallel()

	type input struct {
		UserID uint
		Title  string
		Body   string
		Data   map[string]string
	}
	type expected struct {
		success     bool
		errorCode   int
		userCalls   int
		senderCalls int
	}
	type testCase struct {
		name        string
		input       input
		expected    expected
		userOut     *userSvc.GetUserFcmTokenOut
		senderError error
	}

	defaultUserOut := func() *userSvc.GetUserFcmTokenOut {
		return &userSvc.GetUserFcmTokenOut{
			BaseOut:   dto.BaseOut{Success: true},
			FcmTokens: []*model.FcmToken{{FcmToken: "device-token"}},
		}
	}
	tests := []testCase{
		{
			name: "sends notification to user tokens",
			input: input{
				UserID: 10,
				Title:  "Hello",
				Body:   "World",
				Data:   map[string]string{"key": "value"},
			},
			expected: expected{success: true, userCalls: 1, senderCalls: 1},
			userOut:  defaultUserOut(),
		},
		{
			name:     "rejects empty title",
			input:    input{UserID: 10, Body: "World"},
			expected: expected{errorCode: http.StatusUnprocessableEntity},
			userOut:  defaultUserOut(),
		},
		{
			name:     "rejects empty body",
			input:    input{UserID: 10, Title: "Hello"},
			expected: expected{errorCode: http.StatusUnprocessableEntity},
			userOut:  defaultUserOut(),
		},
		{
			name:     "rejects empty user ID",
			input:    input{Title: "Hello", Body: "World"},
			expected: expected{errorCode: http.StatusUnprocessableEntity},
			userOut:  defaultUserOut(),
		},
		{
			name:     "returns user service failure",
			input:    input{UserID: 10, Title: "Hello", Body: "World"},
			expected: expected{errorCode: http.StatusInternalServerError, userCalls: 1},
			userOut: fixture.Modify(defaultUserOut, func(out *userSvc.GetUserFcmTokenOut) {
				out.Success = false
				out.ErrorCode = http.StatusInternalServerError
				out.ErrorMessage = "token lookup failed"
			}),
		},
		{
			name:        "succeeds when sending fails",
			input:       input{UserID: 10, Title: "Hello", Body: "World"},
			expected:    expected{success: true, userCalls: 1, senderCalls: 1},
			userOut:     defaultUserOut(),
			senderError: errors.New("temporary firebase error"),
		},
		{
			name:     "succeeds when user has no tokens",
			input:    input{UserID: 10, Title: "Hello", Body: "World"},
			expected: expected{success: true, userCalls: 1},
			userOut:  &userSvc.GetUserFcmTokenOut{BaseOut: dto.BaseOut{Success: true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userMock := &mocks.UserServiceMock{
				GetUserFcmTokenFunc: func(ctx context.Context, in *userSvc.GetUserFcmTokenIn) *userSvc.GetUserFcmTokenOut {
					test.AssertEqual(t, "GetUserFcmToken.UserID", tt.input.UserID, in.UserID)
					return tt.userOut
				},
			}
			senderMock := &mocks.SenderMock{
				SendFunc: func(ctx context.Context, title, body string, data map[string]string, token string) (string, error) {
					test.AssertEqual(t, "Sender title", tt.input.Title, title)
					test.AssertEqual(t, "Sender body", tt.input.Body, body)
					test.AssertEqual(t, "Sender data", tt.input.Data, data)
					test.AssertEqual(t, "Sender token", "device-token", token)
					return "message-id", tt.senderError
				},
			}
			service := notificationSvc.NewNotificationService(senderMock, userMock)

			got := service.SendPushNotification(context.Background(), &notificationSvc.SendPushNotificationIn{
				Trace: fixture.TestTrace(), UserID: tt.input.UserID,
				Title: tt.input.Title, Body: tt.input.Body, Data: tt.input.Data,
			})
			test.AssertEqual(t, "SendPushNotification success", tt.expected.success, got.Success)
			test.AssertEqual(t, "SendPushNotification error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "GetUserFcmToken calls", tt.expected.userCalls, len(userMock.GetUserFcmTokenCalls()))
			test.AssertEqual(t, "Firebase Send calls", tt.expected.senderCalls, len(senderMock.SendCalls()))
		})
	}
}
