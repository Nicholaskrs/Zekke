package fixture

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RunMain starts a PostgreSQL container, runs the package's tests, then stops
// the container. Use it as the whole TestMain of any test package that needs
// a database:
//
//	func TestMain(m *testing.M) {
//		os.Exit(fixture.RunMain(m))
//	}
//
// Each test package runs in its own process, so each one gets its own
// container.
func RunMain(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := startPostgres(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start PostgreSQL test container: %v\n", err)
		return 1
	}

	postgresHost, postgresPort, err = containerAddr(container)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get PostgreSQL test container address: %v\n", err)
		_ = container.Terminate(context.Background())
		return 1
	}

	code := m.Run()

	if err := container.Terminate(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "stop PostgreSQL test container: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	return code
}

func startPostgres(ctx context.Context) (testcontainers.Container, error) {
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
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
}

func containerAddr(container testcontainers.Container) (host, port string, err error) {
	ctx := context.Background()

	host, err = container.Host(ctx)
	if err != nil {
		return "", "", fmt.Errorf("host: %w", err)
	}
	mapped, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return "", "", fmt.Errorf("port: %w", err)
	}
	return host, mapped.Port(), nil
}
