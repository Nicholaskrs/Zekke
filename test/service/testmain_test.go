package service_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var postgresHost string
var postgresPort string

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       "postgres",
				"POSTGRES_USER":     "zekke_test",
				"POSTGRES_PASSWORD": "zekke_test",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "start PostgreSQL test container: %v\n", err)
		os.Exit(1)
	}

	postgresHost, err = container.Host(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "get PostgreSQL test container host: %v\n", err)
		_ = container.Terminate(context.Background())
		os.Exit(1)
	}
	port, err := container.MappedPort(context.Background(), "5432/tcp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "get PostgreSQL test container port: %v\n", err)
		_ = container.Terminate(context.Background())
		os.Exit(1)
	}
	postgresPort = port.Port()

	code := m.Run()
	if err := container.Terminate(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "stop PostgreSQL test container: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}
