package rona

import (
	"context"
	"time"
)

// QuickTest constants
const (
	QuickTestValidityDuration = 24 * time.Hour
)

// QuickTestID is a globally unique identifier for each quick test.
type QuickTestID []byte

// QuickTest represents a quick test. The manufacturer enters the list
// of unregistered tests. A test expires 24 hours after it is registered.
type QuickTest struct {
	ID      QuickTestID `json:"id"`
	Expired bool        `json:"expired"`

	// Person is the person's full name that the test is registered to.
	// Person is empty when the test has not been registered yet, or when
	// the test has expired.
	Person string `json:"person,omitempty"`

	CreatedAt    time.Time `json:"created_at"`
	RegisteredAt time.Time `json:"registered_at,omitempty"`
}

// Registered checks if the test has been registered
func (qt *QuickTest) Registered() bool {
	return !qt.RegisteredAt.IsZero()
}

// ShouldExpire checks if the test should expire
func (qt *QuickTest) ShouldExpire() bool {
	return qt.Registered() && time.Since(qt.RegisteredAt) > QuickTestValidityDuration
}

// A QuickTestService interacts with a QuickTest store.
type QuickTestService interface {
	// Retrieve a QuickTest by ID.
	FindQuickTestByID(ctx context.Context, id QuickTestID) (*QuickTest, error)

	// FindQuickTests finds quick tests by the filter criteria.
	FindQuickTests(ctx context.Context, filter *QuickTestFilter) ([]*QuickTest, error)

	// RegisterQuickTest to a given Person.
	// Returns ENOTFOUND if the quick test doesn't exist.
	// Returns EINVALID if the quick test fails validation.
	// Returns EEXPIRED if the quick test has expired.
	RegisterQuickTest(ctx context.Context, reg *QuickTestRegister) error

	// CreateQuickTest creates a new QuickTest.
	// Returns EINVALID if the QuickTest fails to validate.
	CreateQuickTest(ctx context.Context, id QuickTestID) error

	// CreatManyQuickTests creates a batch of new QuickTests.
	//
	// When validation fails, no QuickTests are created.
	// Returns EINVALID if any QuickTest fails to validate.
	CreateManyQuickTests(ctx context.Context, ids []*QuickTestID) error

	// ExpireQuickTest expires the QuickTest.
	// If the quick test is already expired, no action is performed.
	// Returns ENOTFOUND if the quick test does not exist.
	ExpireQuickTestByID(ctx context.Context, id string) error
}

// QuickTestFilter represents the available filters
type QuickTestFilter struct {
	ID      *QuickTestID `json:"id"`
	Expired *bool        `json:"expired"`

	Limit int `json:"limit"`
}

// QuickTestRegister is the set of fields that are needed to register the test.
type QuickTestRegister struct {
	ID     QuickTestID
	Person string
}
