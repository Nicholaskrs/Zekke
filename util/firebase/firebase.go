package firebase

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
	"log"
	"template-go/util/logger"
	"template-go/util/trace"
)

func NewFirebaseClient(serviceAccountKeyPath string) *Client {
	zerologLogger := logger.NewZerologLogger("FirebaseClient")
	trc := &trace.Trace{TraceId: "FirebaseClient"}
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
	logger logger.Logger
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
