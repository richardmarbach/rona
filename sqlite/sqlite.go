package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	// Use sqlite driver
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/richardmarbach/rona"
)

// Database metrics
var (
	testCountGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rona_db_tests",
		Help: "Total number of tests",
	})

	registeredCountGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rona_db_registered",
		Help: "Number of registered tests",
	})

	availableCountGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rona_db_available",
		Help: "Number of available tests",
	})
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Now wraps time.Now. This let's us set the transaction
// time in testing.
var Now = time.Now

// DB represents a database connection.
type DB struct {
	db     *sql.DB
	ctx    context.Context
	cancel func()

	DSN string
}

// NewDB creates a new database connection
func NewDB(dsn string) *DB {
	db := &DB{
		DSN: dsn,
	}
	db.ctx, db.cancel = context.WithCancel(context.Background())
	return db
}

// Open a new database connection
func (db *DB) Open() (err error) {
	if db.DSN == "" {
		return fmt.Errorf("sqlite: DSN is required")
	}

	if db.DSN != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(db.DSN), 0700); err != nil {
			return err
		}
	}

	if db.db, err = sql.Open("sqlite3", db.DSN); err != nil {
		return err
	}

	if _, err := db.db.Exec("PRAGMA journal_mode=wal;"); err != nil {
		return err
	}

	if err := db.migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	go db.monitor()

	return nil
}

// migrate sets up migration tracking and runs the migrations.
func (db *DB) migrate() error {
	if _, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY);`); err != nil {
		return err
	}

	names, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)

	for _, name := range names {
		if err := db.migrateFile(name); err != nil {
			return fmt.Errorf("migration error: name=%q error=%v", name, err)
		}
	}

	return nil
}

func (db *DB) migrateFile(name string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM migrations WHERE name = ?`, name).Scan(&n); err != nil {
		return err
	} else if n != 0 {
		// migration already run
		return nil
	}

	// Run the migration
	if buf, err := fs.ReadFile(migrationsFS, name); err != nil {
		return err
	} else if _, err := tx.Exec(string(buf)); err != nil {
		return err
	}

	if _, err := tx.Exec(`INSERT INTO migrations VALUES (?)`, name); err != nil {
		return err
	}

	return tx.Commit()
}

// Close the database connection
func (db *DB) Close() error {
	db.cancel()

	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:  tx,
		db:  db,
		Now: Now(),
	}, nil
}

// monitor gathers database metrics
func (db *DB) monitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-db.ctx.Done():
			return
		case <-ticker.C:
		}

		if err := db.updateStats(db.ctx); err != nil {
			log.Printf("stats error: %v", err)
		}
	}
}

// updateStats updates the database metrics
func (db *DB) updateStats(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM quick_tests;`).Scan(&n); err != nil {
		return err
	}
	testCountGauge.Set(float64(n))

	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM quick_tests WHERE expired = 0 AND registered_at IS NOT NULL;`).Scan(&n); err != nil {
		return err
	}
	registeredCountGauge.Set(float64(n))

	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM quick_tests WHERE expired = 0 AND registered_at IS NULL;`).Scan(&n); err != nil {
		return err
	}
	availableCountGauge.Set(float64(n))

	return nil
}

// Tx wraps sql.Tx and tracks transaction start time
type Tx struct {
	*sql.Tx
	db  *DB
	Now time.Time
}

// FormatError wraps the sqlite error as an application error when possible.
func FormatError(err error) error {
	if err == nil {
		return nil
	}

	if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
		return rona.Errorf(rona.ECONFLICT, "duplicate record")
	}
	return err
}

// NullString maps empty string to nil
type NullString string

// Scan reads a string value from the database
func (s *NullString) Scan(value interface{}) error {
	if value == nil {
		*(*string)(s) = ""
		return nil
	} else if value, ok := value.(string); ok {
		*(*string)(s) = value
		return nil
	}

	return fmt.Errorf("NullString: cannot scan string: %v", s)
}

// Value formats the string for the database.
func (s *NullString) Value() (driver.Value, error) {
	if string(*s) == "" {
		return nil, nil
	}
	return string(*s), nil
}

// NullTime encodes time as an RFC3339 encoded string.
type NullTime time.Time

// Scan reads a time.Time from the database string
func (n *NullTime) Scan(value interface{}) (err error) {
	if value == nil {
		*(*time.Time)(n) = time.Time{}
		return nil
	} else if value, ok := value.(string); ok {
		*(*time.Time)(n), err = time.Parse(time.RFC3339, value)
		return err
	}

	return fmt.Errorf("NullTime: cannot scan time.Time: %T", value)
}

// Value encodes a time.Time as a database string
func (n *NullTime) Value() (driver.Value, error) {
	if n == nil || (*time.Time)(n).IsZero() {
		return nil, nil
	}
	return (*time.Time)(n).UTC().Format(time.RFC3339), nil
}
