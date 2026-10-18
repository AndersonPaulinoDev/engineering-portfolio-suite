package httpapi

import (
	"context"
	"encoding/json"
	"github.com/AndersonPaulinoDev/opstrack/api/internal/auth"
	"github.com/AndersonPaulinoDev/opstrack/api/internal/domain"
	"github.com/AndersonPaulinoDev/opstrack/api/internal/service"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	auth *auth.Service
	svc  *service.Service
}
type key int

const userKey key = 1
const roleKey key = 2

func New(a *auth.Service, s *service.Service) *Server { return &Server{a, s} }
func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { ok(w, map[string]string{"status": "ok"}) })
	m.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) { ok(w, map[string]string{"status": "ready"}) })
	m.HandleFunc("POST /api/v1/auth/login", s.login)
	m.Handle("GET /api/v1/work-orders", s.guard(http.HandlerFunc(s.list)))
	m.Handle("GET /api/v1/work-orders/{id}", s.guard(http.HandlerFunc(s.get)))
	m.Handle("POST /api/v1/work-orders/{id}/transition", s.guard(http.HandlerFunc(s.transition)))
	m.Handle("POST /api/v1/work-orders/{id}/notes", s.guard(http.HandlerFunc(s.note)))
	m.Handle("POST /api/v1/work-orders/{id}/evidence", s.guard(http.HandlerFunc(s.evidence)))
	m.Handle("GET /api/v1/equipment/{tag}/work-orders", s.guard(http.HandlerFunc(s.equipment)))
	m.Handle("GET /api/v1/audit", s.guard(http.HandlerFunc(s.audit)))
	m.Handle("POST /api/v1/sync", s.guard(http.HandlerFunc(s.sync)))
	return cors(logged(m))
}
func (s *Server) guard(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		u, role, err := s.auth.Verify(tok)
		if err != nil {
			fail(w, 401, "unauthorized", err.Error())
			return
		}
		ctx := context.WithValue(r.Context(), userKey, u)
		ctx = context.WithValue(ctx, roleKey, role)
		n.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !dec(w, r, &in) {
		return
	}
	u, t, err := s.auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		fail(w, 401, "invalid_credentials", err.Error())
		return
	}
	ok(w, map[string]any{"user": u, "access_token": t})
}
func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.List(r.Context(), r.Context().Value(userKey).(string), r.Context().Value(roleKey).(domain.Role))
	if err != nil {
		internal(w, err)
		return
	}
	ok(w, rows)
}
func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, "not_found", err.Error())
		return
	}
	ok(w, x)
}
func (s *Server) transition(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status domain.Status `json:"status"`
	}
	if !dec(w, r, &in) {
		return
	}
	x, err := s.svc.Transition(r.Context(), r.Context().Value(userKey).(string), r.Context().Value(roleKey).(domain.Role), r.PathValue("id"), in.Status)
	if err != nil {
		fail(w, 409, "invalid_transition", err.Error())
		return
	}
	ok(w, x)
}
func (s *Server) note(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Note string `json:"note"`
	}
	if !dec(w, r, &in) {
		return
	}
	x, err := s.svc.AddNote(r.Context(), r.Context().Value(userKey).(string), r.PathValue("id"), in.Note)
	if err != nil {
		fail(w, 400, "invalid_note", err.Error())
		return
	}
	ok(w, x)
}
func (s *Server) evidence(w http.ResponseWriter, r *http.Request) {
	var in struct {
		URI string `json:"uri"`
	}
	if !dec(w, r, &in) {
		return
	}
	x, err := s.svc.AddEvidence(r.Context(), r.Context().Value(userKey).(string), r.PathValue("id"), in.URI)
	if err != nil {
		fail(w, 400, "invalid_evidence", err.Error())
		return
	}
	ok(w, x)
}
func (s *Server) equipment(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ByEquipment(r.Context(), r.PathValue("tag"))
	if err != nil {
		internal(w, err)
		return
	}
	ok(w, rows)
}
func (s *Server) audit(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.Audit(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	ok(w, rows)
}
func (s *Server) sync(w http.ResponseWriter, r *http.Request) {
	var ops []domain.SyncOperation
	if !dec(w, r, &ops) {
		return
	}
	ok(w, s.svc.Sync(r.Context(), r.Context().Value(userKey).(string), r.Context().Value(roleKey).(domain.Role), ops))
}
func dec(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		fail(w, 400, "invalid_json", err.Error())
		return false
	}
	return true
}
func ok(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": data, "error": nil})
}
func fail(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"data": nil, "error": map[string]string{"code": code, "message": msg}})
}

type lw struct {
	http.ResponseWriter
	status int
}

func (x *lw) WriteHeader(c int) { x.status = c; x.ResponseWriter.WriteHeader(c) }
func logged(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		x := &lw{w, 200}
		n.ServeHTTP(x, r)
		log.Printf(`{"method":%q,"path":%q,"status":%d,"duration_ms":%d}`, r.Method, r.URL.Path, x.status, time.Since(start).Milliseconds())
	})
}
func cors(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		n.ServeHTTP(w, r)
	})
}

func internal(w http.ResponseWriter, err error) {
	fail(w, http.StatusInternalServerError, "internal_error", err.Error())
}
