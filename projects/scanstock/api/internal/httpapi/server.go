package httpapi

import (
	"context"
	"encoding/json"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/auth"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/domain"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/service"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	auth *auth.Service
	svc  *service.Service
}
type ctxKey int

const userKey ctxKey = 1

func New(a *auth.Service, s *service.Service) *Server { return &Server{a, s} }
func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { ok(w, map[string]string{"status": "ok"}) })
	m.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) { ok(w, map[string]string{"status": "ready"}) })
	m.HandleFunc("POST /api/v1/auth/login", s.login)
	m.Handle("GET /api/v1/products", s.guard(http.HandlerFunc(s.products)))
	m.Handle("POST /api/v1/products", s.guard(http.HandlerFunc(s.createProduct)))
	m.Handle("GET /api/v1/products/code/{code}", s.guard(http.HandlerFunc(s.byCode)))
	m.Handle("GET /api/v1/transactions", s.guard(http.HandlerFunc(s.transactions)))
	m.Handle("POST /api/v1/transactions", s.guard(http.HandlerFunc(s.record)))
	m.Handle("POST /api/v1/sync", s.guard(http.HandlerFunc(s.sync)))
	return cors(logged(m))
}
func (s *Server) guard(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		u, _, err := s.auth.Verify(tok)
		if err != nil {
			fail(w, 401, "unauthorized", err.Error())
			return
		}
		n.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, u)))
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
	u, t, e := s.auth.Login(r.Context(), in.Email, in.Password)
	if e != nil {
		fail(w, 401, "invalid_credentials", e.Error())
		return
	}
	ok(w, map[string]any{"user": u, "access_token": t})
}
func (s *Server) products(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.Products(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	ok(w, rows)
}
func (s *Server) createProduct(w http.ResponseWriter, r *http.Request) {
	var p domain.Product
	if !dec(w, r, &p) {
		return
	}
	x, e := s.svc.CreateProduct(r.Context(), p)
	if e != nil {
		fail(w, 400, "invalid_product", e.Error())
		return
	}
	ok(w, x)
}
func (s *Server) byCode(w http.ResponseWriter, r *http.Request) {
	x, e := s.svc.FindByCode(r.Context(), r.PathValue("code"))
	if e != nil {
		fail(w, 404, "not_found", e.Error())
		return
	}
	ok(w, x)
}
func (s *Server) transactions(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.Transactions(r.Context(), r.URL.Query().Get("product_id"))
	if err != nil {
		internal(w, err)
		return
	}
	ok(w, rows)
}
func (s *Server) record(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ProductID      string                 `json:"product_id"`
		QuantityChange int                    `json:"quantity_change"`
		Type           domain.TransactionType `json:"transaction_type"`
		DeviceID       string                 `json:"device_id"`
		Reference      string                 `json:"reference"`
		IdempotencyKey string                 `json:"idempotency_key"`
	}
	if !dec(w, r, &in) {
		return
	}
	x, e := s.svc.Record(r.Context(), r.Context().Value(userKey).(string), in.ProductID, in.QuantityChange, in.Type, in.DeviceID, in.Reference, in.IdempotencyKey)
	if e != nil {
		fail(w, 409, "transaction_rejected", e.Error())
		return
	}
	ok(w, x)
}
func (s *Server) sync(w http.ResponseWriter, r *http.Request) {
	var ops []domain.SyncOperation
	if !dec(w, r, &ops) {
		return
	}
	ok(w, s.svc.Sync(r.Context(), r.Context().Value(userKey).(string), ops))
}
func dec(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		fail(w, 400, "invalid_json", e.Error())
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
		st := time.Now()
		x := &lw{w, 200}
		n.ServeHTTP(x, r)
		log.Printf(`{"method":%q,"path":%q,"status":%d,"duration_ms":%d}`, r.Method, r.URL.Path, x.status, time.Since(st).Milliseconds())
	})
}
func cors(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key")
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
