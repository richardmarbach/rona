package mock

import (
	"context"
	"time"

	"github.com/richardmarbach/rona"
)

// QuickTestService mock
type QuickTestService struct {
	FindQuickTestByIDFn        func(ctx context.Context, id rona.QuickTestID) (*rona.QuickTest, error)
	RegisterQuickTestFn        func(ctx context.Context, reg *rona.QuickTestRegister) (*rona.QuickTest, error)
	CreateQuickTestFn          func(ctx context.Context, id rona.QuickTestID) (*rona.QuickTest, error)
	CreateManyQuickTestsFn     func(ctx context.Context, ids []rona.QuickTestID) ([]*rona.QuickTest, error)
	ExpireQuickTestFn          func(ctx context.Context, id rona.QuickTestID) error
	ExpireOutdatedQuickTestsFn func(ctx context.Context, d time.Duration) error
}

func (s *QuickTestService) FindQuickTestByID(ctx context.Context, id rona.QuickTestID) (*rona.QuickTest, error) {
	return s.FindQuickTestByIDFn(ctx, id)
}

func (s *QuickTestService) RegisterQuickTest(ctx context.Context, reg *rona.QuickTestRegister) (*rona.QuickTest, error) {
	return s.RegisterQuickTestFn(ctx, reg)
}

func (s *QuickTestService) CreateQuickTest(ctx context.Context, id rona.QuickTestID) (*rona.QuickTest, error) {
	return s.CreateQuickTestFn(ctx, id)
}

func (s *QuickTestService) CreateManyQuickTests(ctx context.Context, ids []rona.QuickTestID) ([]*rona.QuickTest, error) {
	return s.CreateManyQuickTestsFn(ctx, ids)
}

func (s *QuickTestService) ExpireQuickTest(ctx context.Context, id rona.QuickTestID) error {
	return s.ExpireQuickTestFn(ctx, id)
}

func (s *QuickTestService) ExpireOutdatedQuickTests(ctx context.Context, d time.Duration) error {
	return s.ExpireOutdatedQuickTestsFn(ctx, d)
}
