package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"personal/gosketch/domain"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Server struct {
	r *chi.Mux
	s Service
}

type operations struct {
	Skip bool `json:"skip"`
}

type Service interface {
	ListSessions(ctx context.Context) ([]domain.Session, error)
	CreateSession(ctx context.Context, s domain.Session) (domain.Session, error)
	GetSession(ctx context.Context, id uuid.UUID) (domain.Session, error)
	SkipPhoto(ctx context.Context, id uuid.UUID, skip bool) (domain.Photo, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
}

func NewServer(srv Service) *Server {
	s := &Server{
		r: chi.NewRouter(),
		s: srv,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.r.Route("/sessions", func(r chi.Router) {
		r.Get("/", s.listSessions)
		r.Post("/", s.createSession)

		// Subrouters:
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.getSession)
			r.Delete("/", s.deleteSession)

			r.Route("/photos", func(r chi.Router) {
				r.Post("/", s.createPhoto)
			})
		})
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.r.ServeHTTP(w, req)
}

func (s *Server) listSessions(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sessions, err := s.s.ListSessions(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sessions); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusNoContent)
		log.Printf("Cannot list sessions: %v", err)
		return
	}
}

func (s *Server) createSession(w http.ResponseWriter, req *http.Request) {
	var i domain.Session
	if err := json.NewDecoder(req.Body).Decode(&i); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusNoContent)
		log.Printf("Cannot create session: %v", err)
		return
	}

	ctx := req.Context()
	session, err := s.s.CreateSession(ctx, i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusNoContent)
		log.Printf("Cannot show created session: %v", err)
		return
	}
}

func (s *Server) getSession(w http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(chi.URLParam(req, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		fmt.Println(err)
		return
	}

	ctx := req.Context()
	session, err := s.s.GetSession(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusNoContent)
		log.Printf("Cannot get session: %v", err)
		return
	}
}

func (s *Server) createPhoto(w http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(chi.URLParam(req, "id"))
	if err != nil {
		fmt.Println(err)
		return
	}

	op := operations{
		Skip: false,
	}

	if err := json.NewDecoder(req.Body).Decode(&op); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusNoContent)
		log.Printf("Cannot parse skip: %v", err)
		return
	}

	ctx := req.Context()
	photo, err := s.s.SkipPhoto(ctx, id, op.Skip)
	if err != nil {
		if err.Error() == "session ended" {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(photo); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusNoContent)
		log.Printf("Cannot show photo: %v", err)
		return
	}
}

func (s *Server) deleteSession(w http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(chi.URLParam(req, "id"))
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := req.Context()
	err = s.s.DeleteSession(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "item deleted")
}
