package rona

import (
	"errors"
	"fmt"
)

// Available error codes
const (
	EINTERNAL     = "internal"
	EINVALID      = "invalid"
	ENOTFOUND     = "not_found"
	EEXPIRED      = "expired"
	EUNAUTHORIZED = "unauthorized"
)

// An Error in the application. All non-application
// errors are mapped to EINTERNAL.
type Error struct {
	// Machine readable error code
	Code string

	// Human readable error message
	Message string
}

// Error implements the Error interface
func (e *Error) Error() string {
	return fmt.Sprintf("rona error: code=%s message=%s", e.Code, e.Message)
}

// ErrorCode unwraps the error into an application error code
func ErrorCode(err error) string {
	var e *Error

	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Code
	}
	return EINTERNAL
}
