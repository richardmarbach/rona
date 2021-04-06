package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/richardmarbach/rona"
	ronahttp "github.com/richardmarbach/rona/http"
	"github.com/richardmarbach/rona/mock"
)

func TestGETQuickTest(t *testing.T) {
	// t.Run("get test a's owner", func(t *testing.T) {
	// 	server := MustCreateServer(t)

	// 	server.QuickTestService.FindQuickTestByIDFn = func(ctx context.Context, id rona.QuickTestID) (*rona.QuickTest, error) {
	// 		return &rona.QuickTest{ID: id, Person: "Steve"}, nil
	// 	}

	// 	request, _ := http.NewRequest(http.MethodGet, "/tests/a", nil)
	// 	response := httptest.NewRecorder()

	// 	server.ServeHTTP(response, request)

	// 	got := response.Body.String()
	// 	want := "Steve"

	// 	if got != want {
	// 		t.Errorf("want %v, got %v", got, want)
	// 	}
	// })

	// t.Run("get test b's owner", func(t *testing.T) {
	// 	server := MustCreateServer(t)

	// 	server.QuickTestService.FindQuickTestByIDFn = func(ctx context.Context, id rona.QuickTestID) (*rona.QuickTest, error) {
	// 		if id == "b" {
	// 			return &rona.QuickTest{ID: id, Person: "Bob"}, nil
	// 		}
	// 		return nil, nil
	// 	}

	// 	request, _ := http.NewRequest(http.MethodGet, "/tests/b", nil)
	// 	response := httptest.NewRecorder()

	// 	server.ServeHTTP(response, request)

	// 	got := response.Body.String()
	// 	want := "Bob"

	// 	if got != want {
	// 		t.Errorf("want %v, got %v", got, want)
	// 	}
	// })

	t.Run("return 404 when no quick test is present", func(t *testing.T) {
		server := MustCreateServer(t)

		server.QuickTestService.FindQuickTestByIDFn = func(ctx context.Context, id rona.QuickTestID) (*rona.QuickTest, error) {
			return nil, rona.Errorf(rona.ENOTFOUND, "No id found")
		}

		request, _ := http.NewRequest(http.MethodGet, "/tests/b", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Code
		want := http.StatusNotFound

		if got != want {
			t.Errorf("want %v, got %v", got, want)
		}
	})
}

type Server struct {
	*ronahttp.Server

	QuickTestService mock.QuickTestService
}

func MustCreateServer(tb testing.TB) *Server {
	tb.Helper()

	s := &Server{}
	if server, err := ronahttp.NewServer(&s.QuickTestService); err != nil {
		tb.Fatal(err)
	} else {
		s.Server = server
	}

	return s
}
