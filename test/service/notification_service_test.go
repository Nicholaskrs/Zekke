package service_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	firebaseApp "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	dto "template-go/base/base"
	"template-go/data/model"
	notificationSvc "template-go/modules/notification/svc"
	userSvc "template-go/modules/user/svc"
	"template-go/util/firebase"
)

type notificationExpected struct {
	success     bool
	errorCode   int
	userCalls   int
	sendCalls   int
	deleteCalls int
	sendInputs  []notificationSendInput
}

type notificationSendInput struct {
	title string
	body  string
	data  map[string]string
	token string
}

type notificationSendResult struct {
	messageID string
	err       error
}

type notificationTestCase struct {
	name                     string
	input                    *notificationSvc.SendPushNotificationIn
	expected                 notificationExpected
	GetUserFcmTokenMock      *userSvc.GetUserFcmTokenOut
	DeleteFcmTokenBulkMock   *userSvc.DeleteFcmTokenBulkOut
	deleteFcmTokenMockTokens []string
	SendMock                 map[string]notificationSendResult
	sender                   func(t *testing.T) firebase.Sender
}

func TestNotificationService_SendPushNotification(t *testing.T) {
	t.Parallel()

	tests := []notificationTestCase{
		{
			name: "validates required title before dependencies",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 1, Body: "body",
			},
			expected: notificationExpected{errorCode: http.StatusUnprocessableEntity},
		},
		{
			name: "validates required body before dependencies",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 1, Title: "title",
			},
			expected: notificationExpected{errorCode: http.StatusUnprocessableEntity},
		},
		{
			name: "validates required user ID before dependencies",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), Title: "title", Body: "body",
			},
			expected: notificationExpected{errorCode: http.StatusUnprocessableEntity},
		},
		{
			name: "propagates user token lookup failure",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 4, Title: "title", Body: "body",
			},
			GetUserFcmTokenMock: &userSvc.GetUserFcmTokenOut{
				BaseOut: dto.BaseOut{ErrorCode: http.StatusNotFound, ErrorMessage: "user not found"},
			},
			expected: notificationExpected{errorCode: http.StatusNotFound, userCalls: 1},
		},
		{
			name: "succeeds when user has no tokens",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 4, Title: "title", Body: "body",
			},
			GetUserFcmTokenMock: &userSvc.GetUserFcmTokenOut{
				BaseOut: dto.BaseOut{Success: true},
			},
			expected: notificationExpected{success: true, userCalls: 1},
		},
		{
			name: "sends to every token and treats transient send errors as best effort",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 4, Title: "title", Body: "body", Data: map[string]string{"kind": "notice"},
			},
			GetUserFcmTokenMock: &userSvc.GetUserFcmTokenOut{
				BaseOut: dto.BaseOut{Success: true},
				FcmTokens: []*model.FcmToken{
					{FcmToken: "token-1"},
					{FcmToken: "token-2"},
				},
			},
			SendMock: map[string]notificationSendResult{
				"token-1": {err: errors.New("temporary send error")},
				"token-2": {messageID: "message-id"},
			},
			expected: notificationExpected{
				success: true, userCalls: 1, sendCalls: 2,
				sendInputs: []notificationSendInput{
					{title: "title", body: "body", data: map[string]string{"kind": "notice"}, token: "token-1"},
					{title: "title", body: "body", data: map[string]string{"kind": "notice"}, token: "token-2"},
				},
			},
		},
		{
			name: "removes unregistered tokens",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 4, Title: "title", Body: "body",
			},
			GetUserFcmTokenMock: &userSvc.GetUserFcmTokenOut{
				BaseOut:   dto.BaseOut{Success: true},
				FcmTokens: []*model.FcmToken{{FcmToken: "unregistered-token"}},
			},
			DeleteFcmTokenBulkMock:   &userSvc.DeleteFcmTokenBulkOut{BaseOut: dto.BaseOut{Success: true}},
			deleteFcmTokenMockTokens: []string{"unregistered-token"},
			sender:                   newFirebaseSDKTestSender,
			expected: notificationExpected{
				success: true, userCalls: 1, deleteCalls: 1,
				sendInputs: []notificationSendInput{
					{title: "title", body: "body", token: "unregistered-token"},
				},
			},
		},
		{
			name: "propagates failure to remove unregistered tokens",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 4, Title: "title", Body: "body",
			},
			GetUserFcmTokenMock: &userSvc.GetUserFcmTokenOut{
				BaseOut:   dto.BaseOut{Success: true},
				FcmTokens: []*model.FcmToken{{FcmToken: "unregistered-token"}},
			},
			DeleteFcmTokenBulkMock: &userSvc.DeleteFcmTokenBulkOut{
				BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "cleanup unavailable"},
			},
			deleteFcmTokenMockTokens: []string{"unregistered-token"},
			sender:                   newFirebaseSDKTestSender,
			expected: notificationExpected{
				errorCode: http.StatusServiceUnavailable, userCalls: 1, deleteCalls: 1,
				sendInputs: []notificationSendInput{
					{title: "title", body: "body", token: "unregistered-token"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newNotificationFixture(t)
			var sdkSender *firebaseSDKTestSender
			f.userMock.GetUserFcmTokenFunc = func(context.Context, *userSvc.GetUserFcmTokenIn) *userSvc.GetUserFcmTokenOut {
				return tt.GetUserFcmTokenMock
			}
			f.userMock.DeleteFcmTokenBulkFunc = func(_ context.Context, in *userSvc.DeleteFcmTokenBulkIn) *userSvc.DeleteFcmTokenBulkOut {
				assertEqual(t, "DeleteFcmTokenBulk() tokens", tt.deleteFcmTokenMockTokens, in.Tokens)
				return tt.DeleteFcmTokenBulkMock
			}
			if tt.SendMock != nil {
				f.firebaseMock.SendFunc = func(ctx context.Context, title, body string, data map[string]string, token string) (string, error) {
					result := tt.SendMock[token]
					return result.messageID, result.err
				}
			}
			if tt.sender != nil {
				sender := tt.sender(t)
				sdkSender, _ = sender.(*firebaseSDKTestSender)
				f.service = notificationSvc.NewNotificationService(sender, f.userMock)
			}
			got := f.service.SendPushNotification(context.Background(), tt.input)
			assertEqual(t, "SendPushNotification() success", tt.expected.success, got.Success)
			assertEqual(t, "SendPushNotification() error code", tt.expected.errorCode, got.ErrorCode)
			assertEqual(t, "GetUserFcmToken calls", tt.expected.userCalls, len(f.userMock.GetUserFcmTokenCalls()))
			sendCalls := f.firebaseMock.SendCalls()
			assertEqual(t, "Firebase Send calls", tt.expected.sendCalls, len(sendCalls))
			assertEqual(t, "DeleteFcmTokenBulk calls", tt.expected.deleteCalls, len(f.userMock.DeleteFcmTokenBulkCalls()))
			var gotSendInputs []notificationSendInput
			if sdkSender != nil {
				gotSendInputs = sdkSender.calls
			} else {
				for _, call := range sendCalls {
					gotSendInputs = append(gotSendInputs, notificationSendInput{
						title: call.Title, body: call.Body, data: call.Data, token: call.Token,
					})
				}
			}
			assertEqual(t, "Firebase Send inputs", tt.expected.sendInputs, gotSendInputs)
		})
	}
}

type firebaseSDKTestSender struct {
	client *messaging.Client
	calls  []notificationSendInput
}

func (sender *firebaseSDKTestSender) Send(ctx context.Context, title, body string, data map[string]string, token string) (string, error) {
	sender.calls = append(sender.calls, notificationSendInput{
		title: title, body: body, data: data, token: token,
	})
	return sender.client.Send(ctx, &messaging.Message{
		Data:  data,
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	})
}

func newFirebaseSDKTestSender(t *testing.T) firebase.Sender {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"status":"NOT_FOUND","message":"token is not registered","details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"UNREGISTERED"}]}}`))
	}))
	t.Cleanup(server.Close)

	ctx := context.Background()
	app, err := firebaseApp.NewApp(ctx, &firebaseApp.Config{ProjectID: "test-project"},
		option.WithEndpoint(server.URL), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("create Firebase test app: %v", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		t.Fatalf("create Firebase test messaging client: %v", err)
	}
	return &firebaseSDKTestSender{client: client}
}
