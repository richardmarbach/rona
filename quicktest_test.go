package rona_test

import (
	"strings"
	"testing"
	"time"

	"github.com/richardmarbach/rona"
)

func TestQuickTestID_Validate(t *testing.T) {
	cases := []struct {
		message string
		id      rona.QuickTestID
		isValid bool
	}{
		{"missing id", rona.QuickTestID(""), false},
		{"invalid id", rona.QuickTestID("abcdef"), false},
		{"valid id", rona.NewQuickTestID(), true},
	}

	for _, tc := range cases {
		t.Run(tc.message, func(t *testing.T) {
			err := tc.id.Validate()

			if tc.isValid {
				if err != nil {
					t.Errorf("expected id to be valid but got err %v", err)
				}
			} else {
				if err == nil {
					t.Error("expected an error but didn't get one")
				} else if rona.ErrorCode(err) != rona.EINVALID {
					t.Errorf("expected EINVALID but got %v", rona.ErrorCode(err))
				}
			}
		})
	}
}

func TestQuickTestRegister_Validate(t *testing.T) {
	cases := []struct {
		message string
		reg     *rona.QuickTestRegister
		isValid bool
	}{
		{
			message: "missing id",
			reg:     &rona.QuickTestRegister{},
			isValid: false,
		},
		{
			message: "invalid id",
			reg:     &rona.QuickTestRegister{ID: "abc"},
			isValid: false,
		},
		{
			message: "missing person",
			reg:     &rona.QuickTestRegister{ID: rona.NewQuickTestID()},
			isValid: false,
		},
		{
			message: "person too long",
			reg:     &rona.QuickTestRegister{ID: rona.NewQuickTestID(), Person: strings.Repeat("a", rona.QuickTestMaxPersonLen+1)},
			isValid: false,
		},

		{
			message: "valid",
			reg:     &rona.QuickTestRegister{ID: rona.NewQuickTestID(), Person: "Markus"},
			isValid: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.message, func(t *testing.T) {
			err := tc.reg.Validate()

			if tc.isValid {
				if err != nil {
					t.Errorf("reg=%#v err=%v", tc.reg, err)
				}
			} else {
				if err == nil {
					t.Error("expected an error but didn't get one")
				} else if rona.ErrorCode(err) != rona.EINVALID {
					t.Errorf("expected EINVALID but got %v", rona.ErrorCode(err))
				}
			}
		})
	}
}

func TestQuickTest_Registered(t *testing.T) {
	cases := []struct {
		message      string
		qt           *rona.QuickTest
		isRegistered bool
	}{
		{"not yet registered", &rona.QuickTest{}, false},
		{"registered", &rona.QuickTest{RegisteredAt: time.Now()}, true},
	}

	for _, tc := range cases {
		t.Run(tc.message, func(t *testing.T) {
			isRegistered := tc.qt.Registered()

			if tc.isRegistered {
				if !isRegistered {
					t.Errorf("expected the quick test to be registered: %+v", tc.qt)
				}
			} else {
				if isRegistered {
					t.Errorf("expected the quick test to not be registered: %+v", tc.qt)
				}
			}
		})
	}
}

func TestQuickTest_ShouldExpire(t *testing.T) {
	cases := []struct {
		message   string
		qt        *rona.QuickTest
		isExpired bool
	}{
		{"expired", &rona.QuickTest{RegisteredAt: time.Now().Add(-rona.QuickTestValidityDuration - 1)}, true},
		{"not yet expired", &rona.QuickTest{RegisteredAt: time.Now()}, false},
		{"not yet registered", &rona.QuickTest{}, false},
	}

	for _, tc := range cases {
		t.Run(tc.message, func(t *testing.T) {
			isExpired := tc.qt.ShouldExpire()

			if tc.isExpired {
				if !isExpired {
					t.Errorf("expected the quick test to be expired: %+v", tc.qt)
				}
			} else {
				if isExpired {
					t.Errorf("expected the quick test to not be expired: %+v", tc.qt)
				}
			}
		})
	}
}
