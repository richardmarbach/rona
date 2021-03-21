package sqlite_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/richardmarbach/rona/sqlite"
)

var dump = flag.Bool("dump", false, "save work data")

func TestDB(t *testing.T) {
	MustOpenDB(t)
}

func MustOpenDB(tb testing.TB) *sqlite.DB {
	tb.Helper()

	dsn := "file::memory:?cache=shared"

	if *dump {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			tb.Fatal(err)
		}
		dsn = filepath.Join(dir, "db")
		fmt.Println("dump=" + dsn)
	}

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
