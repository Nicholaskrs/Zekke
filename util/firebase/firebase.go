package firebase

import (
	"context"
	"log"
	logger2 "template-go/core/telemetry/logger"
	"template-go/util/logtrace"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

func NewFirebaseClient(serviceAccountKeyPath string) *Client {
	zerologLogger := logger2.NewZerologLogger("FirebaseClient")
	trc := &logtrace.LogTrace{TraceId: "FirebaseClient"}
	opt := option.WithCredentialsFile(serviceAccountKeyPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	// Get a reference to the messaging client
	client, err := app.Messaging(context.Background())
	if err != nil {
		zerologLogger.FatalErr(trc, err).Msg("error getting messaging client")
	}
	return &Client{
		client: client,
		logger: zerologLogger,
	}
}

type Client struct {
	client *messaging.Client
	logger logger2.Logger
}

func (n *Client) Send(ctx context.Context, title string, body string, data map[string]string, token string) (string, error) {
	return n.client.Send(ctx, &messaging.Message{
		Data:  data,
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	})
}
