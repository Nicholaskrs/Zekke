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

func TestNotificationService_SendPushNotification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		input           *notificationSvc.SendPushNotificationIn
		setup           func(f *notificationFixture)
		sender          func(t *testing.T) firebase.Sender
		wantSuccess     bool
		wantCode        int
		wantUserCalls   int
		wantSendCalls   int
		wantDeleteCalls int
	}{
		{
			name: "validates required title before dependencies",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 1, Body: "body",
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "validates required body before dependencies",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 1, Title: "title",
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "validates required user ID before dependencies",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), Title: "title", Body: "body",
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "propagates user token lookup failure",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 4, Title: "title", Body: "body",
			},
			setup: func(f *notificationFixture) {
				f.userMock.GetUserFcmTokenFunc = func(context.Context, *userSvc.GetUserFcmTokenIn) *userSvc.GetUserFcmTokenOut {
					return &userSvc.GetUserFcmTokenOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusNotFound, ErrorMessage: "user not found"}}
				}
			},
			wantCode:      http.StatusNotFound,
			wantUserCalls: 1,
		},
		{
			name: "succeeds when user has no tokens",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 4, Title: "title", Body: "body",
			},
			setup: func(f *notificationFixture) {
				f.userMock.GetUserFcmTokenFunc = func(context.Context, *userSvc.GetUserFcmTokenIn) *userSvc.GetUserFcmTokenOut {
					return &userSvc.GetUserFcmTokenOut{BaseOut: dto.BaseOut{Success: true}}
				}
			},
			wantSuccess:   true,
			wantUserCalls: 1,
		},
		{
			name: "sends to every token and treats transient send errors as best effort",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 4, Title: "title", Body: "body", Data: map[string]string{"kind": "notice"},
			},
			setup: func(f *notificationFixture) {
				f.userMock.GetUserFcmTokenFunc = func(context.Context, *userSvc.GetUserFcmTokenIn) *userSvc.GetUserFcmTokenOut {
					return &userSvc.GetUserFcmTokenOut{
						BaseOut: dto.BaseOut{Success: true},
						FcmTokens: []*model.FcmToken{
							{FcmToken: "token-1"},
							{FcmToken: "token-2"},
						},
					}
				}

				f.firebaseMock.SendFunc = func(_ context.Context, title, body string, data map[string]string, token string) (string, error) {
					if title != "title" || body != "body" || data["kind"] != "notice" {
						t.Errorf("Send() got title=%q body=%q data=%v", title, body, data)
					}
					if token == "token-1" {
						return "", errors.New("temporary send error")
					}
					return "message-id", nil
				}
			},
			wantSuccess:   true,
			wantUserCalls: 1,
			wantSendCalls: 2,
		},
		{
			name: "removes unregistered tokens",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 4, Title: "title", Body: "body",
			},
			setup: func(f *notificationFixture) {
				setupUnregisteredTokenRemoval(t, f, &userSvc.DeleteFcmTokenBulkOut{BaseOut: dto.BaseOut{Success: true}})
			},
			sender:          newFirebaseSDKTestSender,
			wantSuccess:     true,
			wantUserCalls:   1,
			wantDeleteCalls: 1,
		},
		{
			name: "propagates failure to remove unregistered tokens",
			input: &notificationSvc.SendPushNotificationIn{
				Trace: testTrace(), UserID: 4, Title: "title", Body: "body",
			},
			setup: func(f *notificationFixture) {
				setupUnregisteredTokenRemoval(t, f, &userSvc.DeleteFcmTokenBulkOut{
					BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "cleanup unavailable"},
				})
			},
			sender:          newFirebaseSDKTestSender,
			wantCode:        http.StatusServiceUnavailable,
			wantUserCalls:   1,
			wantDeleteCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newNotificationFixture(t)
			if tt.setup != nil {
				tt.setup(f)
			}
			if tt.sender != nil {
				f.service = notificationSvc.NewNotificationService(tt.sender(t), f.userMock)
			}
			got := f.service.SendPushNotification(context.Background(), tt.input)
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("SendPushNotification() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			if calls := len(f.userMock.GetUserFcmTokenCalls()); calls != tt.wantUserCalls {
				t.Fatalf("GetUserFcmToken calls = %d, want %d", calls, tt.wantUserCalls)
			}
			if calls := len(f.firebaseMock.SendCalls()); calls != tt.wantSendCalls {
				t.Fatalf("Firebase Send calls = %d, want %d", calls, tt.wantSendCalls)
			}
			if calls := len(f.userMock.DeleteFcmTokenBulkCalls()); calls != tt.wantDeleteCalls {
				t.Fatalf("DeleteFcmTokenBulk calls = %d, want %d", calls, tt.wantDeleteCalls)
			}
		})
	}
}

func setupUnregisteredTokenRemoval(t *testing.T, f *notificationFixture, deleteResponse *userSvc.DeleteFcmTokenBulkOut) {
	t.Helper()
	f.userMock.GetUserFcmTokenFunc = func(context.Context, *userSvc.GetUserFcmTokenIn) *userSvc.GetUserFcmTokenOut {
		return &userSvc.GetUserFcmTokenOut{
			BaseOut:   dto.BaseOut{Success: true},
			FcmTokens: []*model.FcmToken{{FcmToken: "unregistered-token"}},
		}
	}
	f.userMock.DeleteFcmTokenBulkFunc = func(_ context.Context, in *userSvc.DeleteFcmTokenBulkIn) *userSvc.DeleteFcmTokenBulkOut {
		if len(in.Tokens) != 1 || in.Tokens[0] != "unregistered-token" {
			t.Errorf("DeleteFcmTokenBulk() tokens = %v, want [unregistered-token]", in.Tokens)
		}
		return deleteResponse
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
