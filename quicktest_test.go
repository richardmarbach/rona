package rona_test

import (
	"testing"
	"time"

	"github.com/richardmarbach/rona"
)

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
