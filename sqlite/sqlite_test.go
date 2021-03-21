package sqlite_test

import (
	"testing"

	"github.com/richardmarbach/rona/sqlite"
)

func TestDB(t *testing.T) {
	MustOpenDB(t)
}

func MustOpenDB(tb testing.TB) *sqlite.DB {
	tb.Helper()

	dsn := ":memory:"

	db := sqlite.NewDB(dsn)
	if err := db.Open(); err != nil {
		tb.Fatalf("failed to open db: %v", err)
	}

	tb.Cleanup(func() {
		if err := db.Close(); err != nil {
			tb.Fatalf("failed to close db: %v", err)
		}
	})

	return db
}
