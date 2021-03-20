package rona

import "time"

// QuickTest represents a quick test. The manufacturer enters the list
// of unregistered tests. A test expires 24 hours after it is registered.
type QuickTest struct {
	ID string `json:"id"`

	// Person is the person's full name that the test is registered to.
	// Person is empty when the test has not been registered yet, or when
	// the test has expired.
	Person string `json:"person,omitempty"`

	CreatedAt    time.Time `json:"created_at"`
	RegisteredAt time.Time `json:"registered_at"`
}

// IsRegistered checked if the test has been registered
func (qt *QuickTest) IsRegistered() bool {
	return !qt.RegisteredAt.IsZero()
}
