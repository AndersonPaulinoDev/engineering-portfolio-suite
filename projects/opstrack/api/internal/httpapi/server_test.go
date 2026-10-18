package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AndersonPaulinoDev/opstrack/api/internal/auth"
	"github.com/AndersonPaulinoDev/opstrack/api/internal/service"
	"github.com/AndersonPaulinoDev/opstrack/api/internal/store"
)

func setup(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	st := store.NewMemoryStore()
	a := auth.New("test-secret", st)
	if err := a.SeedDemo(context.Background()); err != nil {
		t.Fatal(err)
	}
	svc := service.New(st)
	if err := svc.SeedDemo(context.Background()); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(New(a, svc).Handler())
	t.Cleanup(server.Close)

	body := bytes.NewBufferString(`{"email":"worker@opstrack.dev","password":"demo1234"}`)
	res, err := http.Post(server.URL+"/api/v1/auth/login", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var parsed struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Data.AccessToken == "" {
		t.Fatal("missing access token")
	}
	return server, parsed.Data.AccessToken
}

func doAuthed(t *testing.T, method, target, token, body string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestWorkOrderHTTPFlow(t *testing.T) {
	server, token := setup(t)
	res := doAuthed(t, http.MethodGet, server.URL+"/api/v1/work-orders", token, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list status=%d", res.StatusCode)
	}
	res.Body.Close()

	res = doAuthed(t, http.MethodPost, server.URL+"/api/v1/work-orders/wo_1001/transition", token, `{"status":"accepted"}`)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("transition status=%d", res.StatusCode)
	}
	res.Body.Close()

	res = doAuthed(t, http.MethodPost, server.URL+"/api/v1/work-orders/wo_1001/notes", token, `{"note":"Filter dirty"}`)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("note status=%d", res.StatusCode)
	}
	res.Body.Close()

	res = doAuthed(t, http.MethodGet, server.URL+"/api/v1/audit", token, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("audit status=%d", res.StatusCode)
	}
	defer res.Body.Close()
	var body struct {
		Data []any `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data) < 2 {
		t.Fatalf("want audit records, got %d", len(body.Data))
	}
}
