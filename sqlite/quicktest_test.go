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
		if found.CreatedAt.IsZero() {
			t.Errorf("expected CreatedAt to not be zero")
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

func TestQuickTest_Register(t *testing.T) {
	t.Run("register a quick test", func(t *testing.T) {
		s := sqlite.NewQuickTestService(MustOpenDB(t))
		ctx := context.Background()
		quicktest := MustCreateQuickTest(ctx, t, s)

		registered, err := s.RegisterQuickTest(ctx, &rona.QuickTestRegister{
			ID:     quicktest.ID,
			Person: "Jimmy Hendricks",
		})

		if err != nil {
			t.Fatal(err)
		}
		if registered == nil {
			t.Fatal("expected a new QuickTest but didn't get one")
		}
		if registered.ID != quicktest.ID {
			t.Errorf("want %s, got %s", quicktest.ID, registered.ID)
		}
		if registered.Person != "Jimmy Hendricks" {
			t.Errorf("want %s, got %s", quicktest.Person, registered.Person)
		}
		if registered.RegisteredAt.IsZero() {
			t.Errorf("expected RegisteredAt to be set")
		}
	})

	t.Run("return ENOTFOUND when there is no such test", func(t *testing.T) {
		s := sqlite.NewQuickTestService(MustOpenDB(t))
		ctx := context.Background()

		_, err := s.RegisterQuickTest(ctx, &rona.QuickTestRegister{
			ID:     rona.NewQuickTestID(),
			Person: "Jimmy Hendricks",
		})

		if err == nil {
			t.Fatal("expected an error but didn't get one")
		} else if rona.ErrorCode(err) != rona.ENOTFOUND {
			t.Fatalf("expected ENOTFOUND but got %v", err)
		}
	})

	t.Run("return ECONFLICT when the test is already registered", func(t *testing.T) {
		s := sqlite.NewQuickTestService(MustOpenDB(t))
		ctx := context.Background()
		quicktest := MustCreatedRegisteredQuickTest(ctx, t, s, "Jimmy Hendricks")

		_, err := s.RegisterQuickTest(ctx, &rona.QuickTestRegister{
			ID:     quicktest.ID,
			Person: "Jimmy Jones",
		})

		if err == nil {
			t.Fatal("expected an error but didn't get one")
		} else if rona.ErrorCode(err) != rona.ECONFLICT {
			t.Fatalf("expected ECONFLICT but got %v", err)
		}
	})

	t.Run("return EEXPIRED when the test has already expired", func(t *testing.T) {
		t.Skip()
	})
}

func TestQuickTestService_Expire(t *testing.T) {
	t.Run("expire a test", func(t *testing.T) {
		s := sqlite.NewQuickTestService(MustOpenDB(t))
		ctx := context.Background()
		quicktest := MustCreateQuickTest(ctx, t, s)

		err := s.ExpireQuickTest(ctx, quicktest.ID)

		if err != nil {
			t.Fatal(err)
		}

		quicktest = MustFindQuickTest(ctx, t, s, quicktest.ID)

		if !quicktest.Expired {
			t.Errorf("expected quicktest to be expired: %#v", quicktest)
		}

		if quicktest.Person != "" {
			t.Errorf("expected quicktest Person to be unset: %v", quicktest.Person)
		}
	})

	t.Run("return ENOTFOUND when quicktest doesn't exist", func(t *testing.T) {
		s := sqlite.NewQuickTestService(MustOpenDB(t))
		ctx := context.Background()
		err := s.ExpireQuickTest(ctx, rona.NewQuickTestID())

		if err == nil {
			t.Errorf("expected an error but didn't get one")
		} else if rona.ErrorCode(err) != rona.ENOTFOUND {
			t.Errorf("expected ENOTFOUND, got %v", err)
		}
	})
}

func MustFindQuickTest(
	ctx context.Context,
	tb testing.TB,
	s *sqlite.QuickTestService,
	id rona.QuickTestID,
) *rona.QuickTest {
	tb.Helper()
	quicktest, err := s.FindQuickTestByID(ctx, id)
	if err != nil {
		tb.Fatalf("failed to fetch quick test %v: %v", id, err)
	}
	return quicktest
}

func MustCreateQuickTest(
	ctx context.Context,
	tb testing.TB,
	s *sqlite.QuickTestService,
) *rona.QuickTest {
	tb.Helper()

	id := rona.NewQuickTestID()
	quicktest, err := s.CreateQuickTest(ctx, id)
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
	return quicktest
}

func MustCreatedRegisteredQuickTest(
	ctx context.Context,
	tb testing.TB,
	s *sqlite.QuickTestService,
	person string,
) *rona.QuickTest {
	tb.Helper()

	quicktest := MustCreateQuickTest(ctx, tb, s)
	quicktest, err := s.RegisterQuickTest(ctx, &rona.QuickTestRegister{
		ID:     quicktest.ID,
		Person: person,
	})
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
	return quicktest
}
