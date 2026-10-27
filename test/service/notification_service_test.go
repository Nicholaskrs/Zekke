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
}

type notificationTestCase struct {
	name                     string
	input                    *notificationSvc.SendPushNotificationIn
	expected                 notificationExpected
	GetUserFcmTokenMock      *userSvc.GetUserFcmTokenOut
	DeleteFcmTokenBulkMock   *userSvc.DeleteFcmTokenBulkOut
	deleteFcmTokenMockTokens []string
	sendMock                 func(*testing.T, context.Context, string, string, map[string]string, string) (string, error)
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
			sendMock: func(t *testing.T, _ context.Context, title, body string, data map[string]string, token string) (string, error) {
				if title != "title" || body != "body" || data["kind"] != "notice" {
					t.Errorf("Send() got title=%q body=%q data=%v", title, body, data)
				}
				if token == "token-1" {
					return "", errors.New("temporary send error")
				}
				return "message-id", nil
			},
			expected: notificationExpected{success: true, userCalls: 1, sendCalls: 2},
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
			expected:                 notificationExpected{success: true, userCalls: 1, deleteCalls: 1},
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
			expected:                 notificationExpected{errorCode: http.StatusServiceUnavailable, userCalls: 1, deleteCalls: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newNotificationFixture(t)
			f.userMock.GetUserFcmTokenFunc = func(context.Context, *userSvc.GetUserFcmTokenIn) *userSvc.GetUserFcmTokenOut {
				return tt.GetUserFcmTokenMock
			}
			f.userMock.DeleteFcmTokenBulkFunc = func(_ context.Context, in *userSvc.DeleteFcmTokenBulkIn) *userSvc.DeleteFcmTokenBulkOut {
				if len(in.Tokens) != len(tt.deleteFcmTokenMockTokens) {
					t.Errorf("DeleteFcmTokenBulk() tokens = %v, want %v", in.Tokens, tt.deleteFcmTokenMockTokens)
				} else {
					for i, token := range in.Tokens {
						if token != tt.deleteFcmTokenMockTokens[i] {
							t.Errorf("DeleteFcmTokenBulk() tokens = %v, want %v", in.Tokens, tt.deleteFcmTokenMockTokens)
							break
						}
					}
				}
				return tt.DeleteFcmTokenBulkMock
			}
			if tt.sendMock != nil {
				f.firebaseMock.SendFunc = func(ctx context.Context, title, body string, data map[string]string, token string) (string, error) {
					return tt.sendMock(t, ctx, title, body, data, token)
				}
			}
			if tt.sender != nil {
				f.service = notificationSvc.NewNotificationService(tt.sender(t), f.userMock)
			}
			got := f.service.SendPushNotification(context.Background(), tt.input)
			if got.Success != tt.expected.success || got.ErrorCode != tt.expected.errorCode {
				t.Fatalf("SendPushNotification() = %+v, want success %t and code %d", got, tt.expected.success, tt.expected.errorCode)
			}
			if calls := len(f.userMock.GetUserFcmTokenCalls()); calls != tt.expected.userCalls {
				t.Fatalf("GetUserFcmToken calls = %d, want %d", calls, tt.expected.userCalls)
			}
			if calls := len(f.firebaseMock.SendCalls()); calls != tt.expected.sendCalls {
				t.Fatalf("Firebase Send calls = %d, want %d", calls, tt.expected.sendCalls)
			}
			if calls := len(f.userMock.DeleteFcmTokenBulkCalls()); calls != tt.expected.deleteCalls {
				t.Fatalf("DeleteFcmTokenBulk calls = %d, want %d", calls, tt.expected.deleteCalls)
			}
		})
	}
}

type firebaseSDKTestSender struct {
	client *messaging.Client
}

func (sender *firebaseSDKTestSender) Send(ctx context.Context, title, body string, data map[string]string, token string) (string, error) {
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
