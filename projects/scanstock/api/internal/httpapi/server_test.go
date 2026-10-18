package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AndersonPaulinoDev/scanstock/api/internal/auth"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/service"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/store"
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

	res, err := http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBufferString(`{"email":"stock@scanstock.dev","password":"demo1234"}`))
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
	return server, parsed.Data.AccessToken
}

func request(t *testing.T, method, target, token, body string) *http.Response {
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

func TestInventoryHTTPAndIdempotency(t *testing.T) {
	server, token := setup(t)
	res := request(t, http.MethodGet, server.URL+"/api/v1/products", token, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("products status=%d", res.StatusCode)
	}
	res.Body.Close()

	payload := `{"product_id":"prd_1001","quantity_change":2,"transaction_type":"received","device_id":"test","reference":"http-test","idempotency_key":"http-idem-1"}`
	res = request(t, http.MethodPost, server.URL+"/api/v1/transactions", token, payload)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("transaction status=%d", res.StatusCode)
	}
	var first struct {
		Data struct {
			ID string `json:"transaction_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&first); err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	res = request(t, http.MethodPost, server.URL+"/api/v1/transactions", token, payload)
	var second struct {
		Data struct {
			ID string `json:"transaction_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&second); err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if first.Data.ID == "" || first.Data.ID != second.Data.ID {
		t.Fatalf("idempotency failed: %q vs %q", first.Data.ID, second.Data.ID)
	}
}
