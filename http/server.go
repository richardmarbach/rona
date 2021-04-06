package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/richardmarbach/rona"
)

// Server is the applications http server
type Server struct {
	QuickTestService rona.QuickTestService
	Router           http.Handler
	TC               TemplateCache
}

// NewServer creates a new http server
func NewServer(quickTestService rona.QuickTestService) (*Server, error) {
	s := &Server{
		QuickTestService: quickTestService,
	}
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(middleware.Timeout(30 * time.Second))

	router.Route("/tests", func(r chi.Router) {
		r.Get("/", s.showTestForm)
		r.Post("/", s.createTest)
		r.Get("/{testID}", s.getTest)
	})

	tc, err := NewTemplateCache()
	if err != nil {
		return nil, err
	}

	s.Router = router
	s.TC = tc

	return s, nil
}

// Start the http server
func (s *Server) Start() error {
	return http.ListenAndServe(":8080", s.Router)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func (s *Server) showTestForm(w http.ResponseWriter, r *http.Request) {
	if err := s.TC.Render(w, "get-test", nil); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (s *Server) createTest(w http.ResponseWriter, r *http.Request) {}

func (s *Server) getTest(w http.ResponseWriter, r *http.Request) {
	testID := chi.URLParam(r, "testID")

	qt, err := s.QuickTestService.FindQuickTestByID(r.Context(), rona.QuickTestID(testID))

	if err != nil {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "%#v", qt)
}
