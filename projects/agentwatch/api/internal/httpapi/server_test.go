package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/realtime"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/service"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/store"
)

func testServer(t *testing.T) (*httptest.Server, *service.AuthService, *service.AgentWatchService) {
	t.Helper()
	st := store.NewMemoryStore()
	hub := realtime.NewHub()
	go hub.Run()
	auth := service.NewAuthService(st, "test-secret", time.Minute, time.Hour)
	svc := service.NewAgentWatchService(st, hub)
	server := httptest.NewServer(NewServer(auth, svc, hub).Handler())
	t.Cleanup(server.Close)
	return server, auth, svc
}

func registerAndToken(t *testing.T, auth *service.AuthService) string {
	t.Helper()
	_, pair, err := auth.Register(context.Background(), "reviewer@example.com", "password123", "developer")
	if err != nil {
		t.Fatal(err)
	}
	return pair.AccessToken
}

func TestProjectsEndpointRequiresAuthAndReturnsData(t *testing.T) {
	server, auth, svc := testServer(t)
	token := registerAndToken(t, auth)
	if _, err := svc.CreateProject(context.Background(), "Portfolio Project"); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/projects", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", res.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", res.StatusCode)
	}
	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("want 1 project, got %d", len(body.Data))
	}
}

func TestRealtimeWebSocketHandshakeAndEventDelivery(t *testing.T) {
	server, auth, svc := testServer(t)
	token := registerAndToken(t, auth)
	parsed, _ := url.Parse(server.URL)
	conn, err := net.Dial("tcp", parsed.Host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	key := "dGhlIHNhbXBsZSBub25jZQ=="
	request := fmt.Sprintf("GET /ws?token=%s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n", url.QueryEscape(token), parsed.Host, key)
	if _, err := io.WriteString(conn, request); err != nil {
		t.Fatal(err)
	}

	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(status, "101 Switching Protocols") {
		t.Fatalf("unexpected websocket status: %s", status)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line == "\r\n" {
			break
		}
	}

	if _, err := svc.CreateProject(context.Background(), "Realtime Test"); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	first, err := reader.ReadByte()
	if err != nil {
		t.Fatal(err)
	}
	if first&0x0f != 0x1 {
		t.Fatalf("want text frame, opcode=%d", first&0x0f)
	}
	second, err := reader.ReadByte()
	if err != nil {
		t.Fatal(err)
	}
	length := int(second & 0x7f)
	if length == 126 {
		a, _ := reader.ReadByte()
		b, _ := reader.ReadByte()
		length = int(a)<<8 | int(b)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), "project.created") {
		t.Fatalf("unexpected payload: %s", payload)
	}
}
