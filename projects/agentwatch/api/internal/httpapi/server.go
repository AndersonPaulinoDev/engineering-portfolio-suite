package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/domain"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/realtime"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/service"
)

type Server struct {
	auth *service.AuthService
	svc  *service.AgentWatchService
	hub  *realtime.Hub
}

func NewServer(a *service.AuthService, s *service.AgentWatchService, h *realtime.Hub) *Server {
	return &Server{auth: a, svc: s, hub: h}
}
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		respond(w, http.StatusOK, map[string]any{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		respond(w, http.StatusOK, map[string]any{"status": "ready"})
	})
	mux.HandleFunc("POST /api/v1/auth/register", s.register)
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
	mux.Handle("GET /api/v1/projects", s.requireAuth(http.HandlerFunc(s.projects)))
	mux.Handle("POST /api/v1/projects", s.requireAuth(http.HandlerFunc(s.createProject)))
	mux.Handle("GET /api/v1/agents", s.requireAuth(http.HandlerFunc(s.agents)))
	mux.Handle("POST /api/v1/agents/heartbeat", s.requireAuth(http.HandlerFunc(s.heartbeat)))
	mux.Handle("GET /api/v1/tasks", s.requireAuth(http.HandlerFunc(s.tasks)))
	mux.Handle("POST /api/v1/tasks", s.requireAuth(http.HandlerFunc(s.createTask)))
	mux.Handle("POST /api/v1/tasks/{id}/status", s.requireAuth(http.HandlerFunc(s.taskStatus)))
	mux.Handle("GET /api/v1/events", s.requireAuth(http.HandlerFunc(s.events)))
	mux.Handle("GET /api/v1/approvals", s.requireAuth(http.HandlerFunc(s.approvals)))
	mux.Handle("POST /api/v1/approvals", s.requireAuth(http.HandlerFunc(s.requestApproval)))
	mux.Handle("POST /api/v1/approvals/{id}/decision", s.requireAuth(http.HandlerFunc(s.decideApproval)))
	mux.HandleFunc("GET /ws", s.websocket)
	return requestLog(cors(mux))
}

type authCtxKey struct{}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			fail(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}
		claims, err := s.auth.Authenticate(token)
		if err != nil {
			fail(w, http.StatusUnauthorized, "unauthorized", err.Error())
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authCtxKey{}, claims.UserID)))
	})
}
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password, Role string }
	if !decode(w, r, &in) {
		return
	}
	u, pair, err := s.auth.Register(r.Context(), in.Email, in.Password, in.Role)
	if err != nil {
		fail(w, http.StatusBadRequest, "register_failed", err.Error())
		return
	}
	respond(w, http.StatusCreated, map[string]any{"user": u, "tokens": pair})
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if !decode(w, r, &in) {
		return
	}
	u, pair, err := s.auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		fail(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
		return
	}
	respond(w, http.StatusOK, map[string]any{"user": u, "tokens": pair})
}
func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decode(w, r, &in) {
		return
	}
	pair, err := s.auth.Refresh(r.Context(), in.RefreshToken)
	if err != nil {
		fail(w, http.StatusUnauthorized, "invalid_refresh_token", err.Error())
		return
	}
	respond(w, http.StatusOK, pair)
}
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	s.auth.Logout(token)
	w.WriteHeader(http.StatusNoContent)
}
func (s *Server) projects(w http.ResponseWriter, r *http.Request) {
	v, err := s.svc.Projects(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	respond(w, http.StatusOK, v)
}
func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name string }
	if !decode(w, r, &in) {
		return
	}
	v, err := s.svc.CreateProject(r.Context(), in.Name)
	if err != nil {
		fail(w, 400, "invalid_project", err.Error())
		return
	}
	respond(w, 201, v)
}
func (s *Server) agents(w http.ResponseWriter, r *http.Request) {
	v, err := s.svc.Agents(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	respond(w, 200, v)
}
func (s *Server) heartbeat(w http.ResponseWriter, r *http.Request) {
	var in struct{ ID, Name, ProjectID, Type, State, Message string }
	if !decode(w, r, &in) {
		return
	}
	v, err := s.svc.Heartbeat(r.Context(), in.ID, in.Name, in.ProjectID, in.Type, domain.AgentState(in.State), in.Message)
	if err != nil {
		fail(w, 400, "invalid_heartbeat", err.Error())
		return
	}
	respond(w, 200, v)
}
func (s *Server) tasks(w http.ResponseWriter, r *http.Request) {
	v, err := s.svc.Tasks(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	respond(w, 200, v)
}
func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var in struct{ ProjectID, AgentID, Title, Description string }
	if !decode(w, r, &in) {
		return
	}
	v, err := s.svc.CreateTask(r.Context(), in.ProjectID, in.AgentID, in.Title, in.Description)
	if err != nil {
		fail(w, 400, "invalid_task", err.Error())
		return
	}
	respond(w, 201, v)
}
func (s *Server) taskStatus(w http.ResponseWriter, r *http.Request) {
	var in struct{ Status string }
	if !decode(w, r, &in) {
		return
	}
	v, err := s.svc.UpdateTaskStatus(r.Context(), r.PathValue("id"), domain.TaskStatus(in.Status))
	if err != nil {
		fail(w, 400, "invalid_transition", err.Error())
		return
	}
	respond(w, 200, v)
}
func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	v, err := s.svc.Events(r.Context(), limit)
	if err != nil {
		internal(w, err)
		return
	}
	respond(w, 200, v)
}
func (s *Server) approvals(w http.ResponseWriter, r *http.Request) {
	v, err := s.svc.Approvals(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	respond(w, 200, v)
}
func (s *Server) requestApproval(w http.ResponseWriter, r *http.Request) {
	var in struct{ ProjectID, AgentID, TaskID, Action, Reason string }
	if !decode(w, r, &in) {
		return
	}
	v, err := s.svc.RequestApproval(r.Context(), in.ProjectID, in.AgentID, in.TaskID, in.Action, in.Reason)
	if err != nil {
		fail(w, 400, "invalid_approval", err.Error())
		return
	}
	respond(w, 201, v)
}
func (s *Server) decideApproval(w http.ResponseWriter, r *http.Request) {
	var in struct{ Decision string }
	if !decode(w, r, &in) {
		return
	}
	v, err := s.svc.DecideApproval(r.Context(), r.PathValue("id"), in.Decision)
	if err != nil {
		fail(w, 400, "invalid_decision", err.Error())
		return
	}
	respond(w, 200, v)
}
func (s *Server) websocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if _, err := s.auth.Authenticate(token); err != nil {
		fail(w, 401, "unauthorized", "invalid websocket token")
		return
	}
	if err := realtime.ServeWebSocket(w, r, s.hub); err != nil {
		log.Printf("websocket: %v", err)
	}
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		fail(w, 400, "invalid_json", err.Error())
		return false
	}
	return true
}
func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data, "error": nil})
}
func fail(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": nil, "error": map[string]string{"code": code, "message": msg}})
}
func internal(w http.ResponseWriter, err error) {
	if errors.Is(err, context.Canceled) {
		return
	}
	log.Printf("internal error: %v", err)
	fail(w, 500, "internal_error", "internal server error")
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type loggingWriter struct {
	http.ResponseWriter
	status int
}

func (w *loggingWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying response writer does not support hijacking")
	}
	return hijacker.Hijack()
}
func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &loggingWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(lw, r)
		log.Printf(`{"method":%q,"path":%q,"status":%d,"duration_ms":%d}`, r.Method, r.URL.Path, lw.status, time.Since(start).Milliseconds())
	})
}
