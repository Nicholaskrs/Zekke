package fixture

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"template-go/util/config"
	"template-go/util/logtrace"
)

// ---- Shared helpers (used by every fixture file) ----

func testConfig(cfg config.Config) config.Config {
	cfg.JwtSecret = "service-test-secret"
	cfg.JwtIssuer = "zekke-service-test"
	return cfg
}

func TestTrace() *logtrace.LogTrace {
	return &logtrace.LogTrace{TraceId: uuid.NewString()}
}

// shortID returns a short random string, handy for unique usernames/emails.
func shortID() string {
	return uuid.NewString()[:8]
}

// mustCreate inserts a row and fails the test if it can't.
// `what` is only used in the error message, e.g. "user", "role".
func mustCreate(t *testing.T, db *gorm.DB, value any, what string) {
	t.Helper()
	if err := db.Create(value).Error; err != nil {
		t.Fatalf("seed %s: %v", what, err)
	}
}

// orDefault returns v, or fallback when v is empty.
// Used by seed helpers so empty row fields mean "use the default".
func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func Modify[T any](create func() T, modify func(T)) T {
	data := create()
	modify(data)
	return data
}
