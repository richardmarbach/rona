package rona_test

import (
	"testing"
	"time"

	"github.com/richardmarbach/rona"
)

func TestQuickTest_IsRegistered(t *testing.T) {
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
			isRegistered := tc.qt.IsRegistered()

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
