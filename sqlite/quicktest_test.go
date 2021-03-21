package sqlite_test

import (
	"context"
	"testing"

	"github.com/richardmarbach/rona"
	"github.com/richardmarbach/rona/sqlite"
)

func TestQuickTestService_FindQuickTestByID(t *testing.T) {
	t.Run("find record by id", func(t *testing.T) {
		s := sqlite.NewQuickTestService(MustOpenDB(t))
		ctx := context.Background()

		quicktest := MustCreateQuickTest(ctx, t, s)

		found, err := s.FindQuickTestByID(ctx, quicktest.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if found.ID != quicktest.ID {
			t.Errorf("want ID %v, got %v", quicktest.ID, found.ID)
		}
	})

	t.Run("no record found", func(t *testing.T) {
		db := MustOpenDB(t)
		s := sqlite.NewQuickTestService(db)

		ctx := context.Background()
		id := rona.NewQuickTestID()

		quicktest, err := s.FindQuickTestByID(ctx, id)
		if err == nil {
			t.Errorf("expected an error but did not get one: %#v", quicktest)
		} else if rona.ErrorCode(err) != rona.ENOTFOUND {
			t.Errorf("expected ENOTFOUND, got %v", err)
		}
	})
}

func TestQuickTestService_CreateQuickTest(t *testing.T) {
	t.Run("create quick test", func(t *testing.T) {
		s := sqlite.NewQuickTestService(MustOpenDB(t))
		ctx := context.Background()

		id := rona.NewQuickTestID()

		quicktest, err := s.CreateQuickTest(ctx, id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if quicktest.ID != id {
			t.Errorf("want id %v, got %v", id, quicktest.ID)
		}
		if quicktest.CreatedAt.IsZero() {
			t.Errorf("expected CreatedAt to be set")
		}
	})

	t.Run("inserting duplicate quicktest fails with ECONFLICT", func(t *testing.T) {
		s := sqlite.NewQuickTestService(MustOpenDB(t))
		ctx := context.Background()

		quicktest := MustCreateQuickTest(ctx, t, s)

		_, err := s.CreateQuickTest(ctx, quicktest.ID)

		if err == nil {
			t.Fatal("want an error but didn't get one")
		} else if rona.ErrorCode(err) != rona.ECONFLICT {
			t.Fatalf("want ECONFLICT, got %v", err)
		}
	})
}

func MustCreateQuickTest(ctx context.Context, tb testing.TB, s *sqlite.QuickTestService) *rona.QuickTest {
	tb.Helper()

	id := rona.NewQuickTestID()
	quicktest, err := s.CreateQuickTest(ctx, id)
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
	return quicktest
}
