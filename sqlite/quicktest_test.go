package sqlite_test

import (
	"context"
	"testing"

	"github.com/richardmarbach/rona"
	"github.com/richardmarbach/rona/sqlite"
)

func TestQuickTestService_FindQuickTestByID(t *testing.T) {
	t.Run("find record by id", func(t *testing.T) {
		ctx, s := createService(t)

		quicktest := MustCreateQuickTest(ctx, t, s)

		found, err := s.FindQuickTestByID(ctx, quicktest.ID)
		assertNoError(t, err)

		if found.ID != quicktest.ID {
			t.Errorf("want ID %v, got %v", quicktest.ID, found.ID)
		}
		if found.CreatedAt.IsZero() {
			t.Errorf("expected CreatedAt to not be zero")
		}
	})

	t.Run("no record found", func(t *testing.T) {
		ctx, s := createService(t)

		id := rona.NewQuickTestID()

		_, err := s.FindQuickTestByID(ctx, id)
		assertErrorCode(t, err, rona.ENOTFOUND)
	})
}

func TestQuickTestService_CreateQuickTest(t *testing.T) {
	t.Run("create quick test", func(t *testing.T) {
		ctx, s := createService(t)

		id := rona.NewQuickTestID()

		quicktest, err := s.CreateQuickTest(ctx, id)
		assertNoError(t, err)

		if quicktest.ID != id {
			t.Errorf("want id %v, got %v", id, quicktest.ID)
		}
		if quicktest.CreatedAt.IsZero() {
			t.Errorf("expected CreatedAt to be set")
		}
	})

	t.Run("inserting duplicate quicktest fails with ECONFLICT", func(t *testing.T) {
		ctx, s := createService(t)
		quicktest := MustCreateQuickTest(ctx, t, s)

		_, err := s.CreateQuickTest(ctx, quicktest.ID)

		assertErrorCode(t, err, rona.ECONFLICT)
	})
}

func TestQuickTestService_CreateMany(t *testing.T) {
	t.Run("create no quick tests", func(t *testing.T) {
		ctx, s := createService(t)

		ids := []rona.QuickTestID{}

		quicktests, err := s.CreateManyQuickTests(ctx, ids)
		assertNoError(t, err)
		if len(quicktests) != 0 {
			t.Errorf("expected 0 quicktest, got %d", len(quicktests))
		}
	})
	t.Run("create one quick test", func(t *testing.T) {
		ctx, s := createService(t)

		ids := []rona.QuickTestID{
			rona.NewQuickTestID(),
		}

		quicktests, err := s.CreateManyQuickTests(ctx, ids)
		assertNoError(t, err)

		if len(quicktests) != 1 {
			t.Errorf("expected 1 quicktest, got %d", len(quicktests))
		}
	})

	t.Run("create many quick tests", func(t *testing.T) {
		ctx, s := createService(t)

		ids := []rona.QuickTestID{
			rona.NewQuickTestID(),
			rona.NewQuickTestID(),
			rona.NewQuickTestID(),
			rona.NewQuickTestID(),
		}

		quicktests, err := s.CreateManyQuickTests(ctx, ids)
		assertNoError(t, err)

		for i, quicktest := range quicktests {
			if quicktest.ID != ids[i] {
				t.Errorf("[%d] wanted id %v, got %v", i, ids[i], quicktest.ID)
			}

			if quicktest.CreatedAt.IsZero() {
				t.Errorf("[%d] expected CreatedAt to not be empty", i)
			}
		}
	})

	t.Run("fails with ECONFLICT if the quicktest already exists", func(t *testing.T) {
		ctx, s := createService(t)

		exists := MustCreateQuickTest(ctx, t, s)

		ids := []rona.QuickTestID{
			rona.NewQuickTestID(),
			exists.ID,
		}

		_, err := s.CreateManyQuickTests(ctx, ids)
		assertErrorCode(t, err, rona.ECONFLICT)

		_, err = s.FindQuickTestByID(ctx, ids[0])
		assertErrorCode(t, err, rona.ENOTFOUND)
	})

	t.Run("fails if any of the IDs fail validation", func(t *testing.T) {
		ctx, s := createService(t)

		ids := []rona.QuickTestID{
			rona.NewQuickTestID(),
			"abc",
		}

		_, err := s.CreateManyQuickTests(ctx, ids)
		assertErrorCode(t, err, rona.EINVALID)
	})
}

func TestQuickTestService_Register(t *testing.T) {
	t.Run("register a quick test", func(t *testing.T) {
		ctx, s := createService(t)
		quicktest := MustCreateQuickTest(ctx, t, s)

		registered, err := s.RegisterQuickTest(ctx, &rona.QuickTestRegister{
			ID:     quicktest.ID,
			Person: "Jimmy Hendricks",
		})
		assertNoError(t, err)

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
		ctx, s := createService(t)

		_, err := s.RegisterQuickTest(ctx, &rona.QuickTestRegister{
			ID:     rona.NewQuickTestID(),
			Person: "Jimmy Hendricks",
		})

		assertErrorCode(t, err, rona.ENOTFOUND)
	})

	t.Run("return ECONFLICT when the test is already registered", func(t *testing.T) {
		ctx, s := createService(t)
		quicktest := MustCreatedRegisteredQuickTest(ctx, t, s, "Jimmy Hendricks")

		_, err := s.RegisterQuickTest(ctx, &rona.QuickTestRegister{
			ID:     quicktest.ID,
			Person: "Jimmy Jones",
		})

		assertErrorCode(t, err, rona.ECONFLICT)
	})

	t.Run("return EEXPIRED when the test has already expired", func(t *testing.T) {
		ctx, s := createService(t)
		quicktest := MustCreateExpiredQuickTest(ctx, t, s)

		_, err := s.RegisterQuickTest(ctx, &rona.QuickTestRegister{
			ID:     quicktest.ID,
			Person: "Jimmy Jones",
		})

		assertErrorCode(t, err, rona.EEXPIRED)
	})
}

func TestQuickTestService_Expire(t *testing.T) {
	t.Run("expire a test", func(t *testing.T) {
		ctx, s := createService(t)
		quicktest := MustCreateQuickTest(ctx, t, s)

		err := s.ExpireQuickTest(ctx, quicktest.ID)
		assertNoError(t, err)

		quicktest = MustFindQuickTest(ctx, t, s, quicktest.ID)

		if !quicktest.Expired {
			t.Errorf("expected quicktest to be expired: %#v", quicktest)
		}

		if quicktest.Person != "" {
			t.Errorf("expected quicktest Person to be unset: %v", quicktest.Person)
		}
	})

	t.Run("return ENOTFOUND when quicktest doesn't exist", func(t *testing.T) {
		ctx, s := createService(t)
		err := s.ExpireQuickTest(ctx, rona.NewQuickTestID())

		assertErrorCode(t, err, rona.ENOTFOUND)
	})
}

func createService(tb testing.TB) (context.Context, *sqlite.QuickTestService) {
	tb.Helper()
	s := sqlite.NewQuickTestService(MustOpenDB(tb))
	return context.Background(), s
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

	assertNoError(tb, err)
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
	assertNoError(tb, err)
	return quicktest
}

func MustCreateExpiredQuickTest(
	ctx context.Context,
	tb testing.TB,
	s *sqlite.QuickTestService,
) *rona.QuickTest {
	tb.Helper()

	quicktest := MustCreateQuickTest(ctx, tb, s)
	err := s.ExpireQuickTest(ctx, quicktest.ID)
	assertNoError(tb, err)
	return MustFindQuickTest(ctx, tb, s, quicktest.ID)
}

func assertNoError(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("expected no error, got %v", err)
	}
}

func assertErrorCode(tb testing.TB, err error, code string) {
	tb.Helper()

	if err == nil {
		tb.Errorf("expected an error but didn't get one")
	} else if rona.ErrorCode(err) != code {
		tb.Errorf("expected %s, got %v", code, err)
	}
}
