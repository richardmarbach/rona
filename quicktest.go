package rona

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// QuickTest constants
const (
	// Aparantly the offial record for the worlds longest first name is 1000
	// characters long. Let's just multiply that by 4 and hope parents don't
	// get it in their head that they need to break that record...
	QuickTestMaxPersonLen     = 4000
	QuickTestValidityDuration = 24 * time.Hour
)

// QuickTestID is a globally unique identifier for each quick test.
type QuickTestID string

// NewQuickTestID generates a new random quicktest ID.
func NewQuickTestID() QuickTestID {
	id := uuid.New()
	return QuickTestID(id.String())
}

// Validate the ID
func (id QuickTestID) Validate() error {
	if id == "" {
		return Errorf(EINVALID, "id is required")
	} else if _, err := uuid.Parse(string(id)); err != nil {
		return Errorf(EINVALID, "invalid id: %v", err)
	}
	return nil
}

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

	// RegisterQuickTest to a given Person.
	// Returns ENOTFOUND if the quick test doesn't exist.
	// Returns EINVALID if the quick test fails validation.
	// Returns EEXPIRED if the quick test has expired.
	RegisterQuickTest(ctx context.Context, reg *QuickTestRegister) (*QuickTest, error)

	// CreateQuickTest creates a new QuickTest.
	// Returns EINVALID if the QuickTest fails to validate.
	// Return ECONFLICT if the QuickTest already exists.
	CreateQuickTest(ctx context.Context, id QuickTestID) (*QuickTest, error)

	// CreatManyQuickTests creates a batch of new QuickTests.
	//
	// When validation fails, no QuickTests are created.
	// Returns EINVALID if any QuickTest fails to validate.
	CreateManyQuickTests(ctx context.Context, ids []*QuickTestID) ([]*QuickTest, error)

	// Expire a QuickTest
	// Returns ENOTFOUND when the quick test doesn't exist.
	ExpireQuickTest(ctx context.Context, id QuickTestID) error

	// ExpireOutdatedQuickTests expires all quick tests older than the
	// given duration.
	ExpireOutdatedQuickTests(ctx context.Context, d time.Duration) error
}

// QuickTestRegister is the set of fields that are needed to register the test.
type QuickTestRegister struct {
	ID     QuickTestID
	Person string
}

// Validate the fields required for registration.
func (r *QuickTestRegister) Validate() error {
	if err := r.ID.Validate(); err != nil {
		return err
	} else if r.Person == "" {
		return Errorf(EINVALID, "name is required")
	} else if len(r.Person) > QuickTestMaxPersonLen {
		return Errorf(EINVALID, "name is too long")
	}
	return nil
}
