package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/richardmarbach/rona"
)

var _ rona.QuickTestService = &QuickTestService{}

// QuickTestService manages quick tests in the sqlite database.
type QuickTestService struct {
	db *DB
}

// NewQuickTestService creates a new QuickTestService
func NewQuickTestService(db *DB) *QuickTestService {
	return &QuickTestService{db: db}
}

// FindQuickTestByID retrieves a quicktest by id.
func (s *QuickTestService) FindQuickTestByID(ctx context.Context, id rona.QuickTestID) (*rona.QuickTest, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRow(`
		SELECT
			id,
			person,
			expired,
			created_at,
			registered_at
		FROM quick_tests
		WHERE id = ?
		LIMIT 1
	`, id)

	var quicktest rona.QuickTest
	if err := row.Scan(
		&quicktest.ID,
		(*NullString)(&quicktest.Person),
		&quicktest.Expired,
		(*NullTime)(&quicktest.CreatedAt),
		(*NullTime)(&quicktest.RegisteredAt),
	); err != nil && err == sql.ErrNoRows {
		return nil, rona.Errorf(rona.ENOTFOUND, "No quick test found for %v", id)
	} else if err != nil {
		return nil, err
	}

	if row.Err() != nil {
		return nil, row.Err()
	}

	return &quicktest, nil
}

// CreateQuickTest creates a new quicktest
func (s *QuickTestService) CreateQuickTest(ctx context.Context, id rona.QuickTestID) (*rona.QuickTest, error) {
	quicktests, err := s.CreateManyQuickTests(ctx, []rona.QuickTestID{id})
	if err != nil {
		return nil, err
	}

	if len(quicktests) != 1 {
		return nil, rona.Errorf(rona.EINTERNAL, "expected quick test to be created, but wasn't")
	}
	return quicktests[0], nil
}

// CreateManyQuickTests creates quick tests in batches.
func (s *QuickTestService) CreateManyQuickTests(ctx context.Context, ids []rona.QuickTestID) ([]*rona.QuickTest, error) {
	for _, id := range ids {
		if err := id.Validate(); err != nil {
			return nil, err
		}
	}

	if len(ids) == 0 {
		return []*rona.QuickTest{}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	quicktests := make([]*rona.QuickTest, 0, len(ids))
	for _, id := range ids {
		quicktest := &rona.QuickTest{
			ID:        id,
			CreatedAt: tx.Now,
		}
		quicktests = append(quicktests, quicktest)
	}

	valueStrings := make([]string, 0, len(quicktests))
	valueArgs := make([]interface{}, 0, len(quicktests)*2)
	for _, quicktest := range quicktests {
		valueStrings = append(valueStrings, "(?, ?)")
		valueArgs = append(valueArgs, quicktest.ID)
		valueArgs = append(valueArgs, (*NullTime)(&quicktest.CreatedAt))
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO quick_tests (id, created_at)
		VALUES %s;
		`, strings.Join(valueStrings, ",")),
		valueArgs...,
	); err != nil {
		return nil, FormatError(err)
	}

	return quicktests, tx.Commit()
}

// RegisterQuickTest registers a new QuickTest
func (s *QuickTestService) RegisterQuickTest(ctx context.Context, reg *rona.QuickTestRegister) (*rona.QuickTest, error) {
	if err := reg.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	quicktest, err := s.FindQuickTestByID(ctx, reg.ID)
	if err != nil {
		return nil, err
	}

	if quicktest.Registered() {
		return nil, rona.Errorf(rona.ECONFLICT, "test has already been registered")
	} else if quicktest.Expired {
		return nil, rona.Errorf(rona.EEXPIRED, "test has already expired")
	}

	quicktest.Person = reg.Person
	quicktest.RegisteredAt = tx.Now

	if _, err := tx.ExecContext(ctx, `
		UPDATE quick_tests
		SET person = ?,
			registered_at = ?
		WHERE id = ?
	`,
		(*NullString)(&quicktest.Person),
		(*NullTime)(&quicktest.RegisteredAt),
		quicktest.ID,
	); err != nil {
		return nil, FormatError(err)
	}

	return quicktest, tx.Commit()
}

// ExpireQuickTest by ID. An expired quicktest removes PII.
func (s *QuickTestService) ExpireQuickTest(ctx context.Context, id rona.QuickTestID) error {
	if err := id.Validate(); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		UPDATE quick_tests
		SET expired = ?,
			person = ?
		WHERE id = ?
	`,
		true,
		"",
		id,
	)
	if err != nil {
		return FormatError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return FormatError(err)
	}
	if rows != 1 {
		return rona.Errorf(rona.ENOTFOUND, "quick test does not exist: %v", id)
	}

	return tx.Commit()
}

// ExpireOutdatedQuickTests expires all quick tests registered after the given duration.
func (s *QuickTestService) ExpireOutdatedQuickTests(ctx context.Context, d time.Duration) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE quick_tests
		SET expired = ?,
			person = ?
		WHERE 
			expired = 0 AND
			registered_at IS NOT NULL AND
			strftime('%s', registered_at) < strftime('%s', DATETIME('now', ?))
	`,
		true,
		"",
		fmt.Sprintf("-%d second", int64(d.Seconds())),
	); err != nil {
		return FormatError(err)
	}

	return tx.Commit()
}
