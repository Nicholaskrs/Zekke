package fixture

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	appDatabase "template-go/core/database"
	"template-go/core/telemetry/logger"
	"template-go/util/config"
)

// Set by RunMain once the test container is up.
var (
	postgresHost string
	postgresPort string
)

// NewTestDatabase creates an isolated, migrated database for one test and
// drops it again when the test finishes.
func NewTestDatabase(t *testing.T) (*gorm.DB, config.Config) {
	t.Helper()

	if postgresHost == "" || postgresPort == "" {
		t.Fatal("test PostgreSQL address not set: use fixture.RunMain in TestMain")
	}

	timezone := os.Getenv("DB_TIMEZONE")
	if timezone == "" {
		timezone = "UTC"
	}
	cfg := config.Config{
		DbUser:         "zekke_test",
		DbPassword:     "zekke_test",
		DbHost:         postgresHost,
		DbPort:         postgresPort,
		DbDriver:       "postgres",
		DbTimezone:     timezone,
		DbMaxOpenConns: 2,
		DbMaxIdleConns: 1,
	}

	// 1. Create an empty database with a unique name.
	adminDB, err := gorm.Open(postgres.Open(postgresDSN(cfg, "postgres")), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect to test PostgreSQL instance: %v", err)
	}
	adminSQLDB, err := adminDB.DB()
	if err != nil {
		t.Fatalf("get test PostgreSQL connection: %v", err)
	}

	cfg.DbName = "zekke_test_" + uuid.NewString()
	if err := adminDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, cfg.DbName)).Error; err != nil {
		_ = adminSQLDB.Close()
		t.Fatalf("create isolated test database: %v", err)
	}

	// 2. Register cleanup: close our connection, then drop the database.
	var testSQLDB *sql.DB
	t.Cleanup(func() {
		if testSQLDB != nil {
			if err := testSQLDB.Close(); err != nil {
				t.Errorf("close isolated test database connection: %v", err)
			}
		}
		dropDatabase(t, cfg)
	})

	if err := adminSQLDB.Close(); err != nil {
		t.Fatalf("close test database administration connection: %v", err)
	}

	// 3. Connect to the new database and apply the production migrations.
	db := appDatabase.NewEngine(&cfg).SetupModels(logger.NewZerologLogger("service-test"), &cfg)
	testSQLDB, err = db.DB()
	if err != nil {
		t.Fatalf("get isolated test database connection: %v", err)
	}

	if err := appDatabase.NewEngine(&cfg).MigrateModels(logger.NewZerologLogger("service-test"), db); err != nil {
		t.Fatalf("apply production migrations to isolated test database: %v", err)
	}
	return db, cfg
}

// dropDatabase reconnects to the admin database and drops cfg.DbName.
func dropDatabase(t *testing.T, cfg config.Config) {
	t.Helper()

	cleanupDB, err := gorm.Open(postgres.Open(postgresDSN(cfg, "postgres")), &gorm.Config{})
	if err != nil {
		t.Errorf("reconnect to test PostgreSQL instance for cleanup: %v", err)
		return
	}
	cleanupSQLDB, err := cleanupDB.DB()
	if err != nil {
		t.Errorf("get test PostgreSQL cleanup connection: %v", err)
		return
	}
	defer cleanupSQLDB.Close()

	if err := cleanupDB.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, cfg.DbName)).Error; err != nil {
		t.Errorf("drop isolated test database: %v", err)
	}
}

func postgresDSN(cfg config.Config, database string) string {
	connectionURL := url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(cfg.DbHost, cfg.DbPort),
		User:   url.UserPassword(cfg.DbUser, cfg.DbPassword),
		Path:   database,
	}
	query := connectionURL.Query()
	query.Set("sslmode", "disable")
	query.Set("TimeZone", cfg.DbTimezone)
	connectionURL.RawQuery = strings.ReplaceAll(query.Encode(), "%2F", "/")
	return connectionURL.String()
}
