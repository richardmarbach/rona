package sqlite

import (
	"context"
	"database/sql"

	"github.com/richardmarbach/rona"
)

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
	if err := id.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	quicktest := &rona.QuickTest{
		ID:        id,
		CreatedAt: tx.Now,
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO quick_tests (id, created_at) 
		VALUES (?, ?);
	`,
		quicktest.ID,
		(*NullTime)(&quicktest.CreatedAt),
	); err != nil {
		return nil, FormatError(err)
	}

	return quicktest, tx.Commit()
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
