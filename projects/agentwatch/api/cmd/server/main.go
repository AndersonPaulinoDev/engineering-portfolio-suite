package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/httpapi"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/realtime"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/service"
)

func main() {
	secret := strings.TrimSpace(os.Getenv("AUTH_SECRET"))
	if len(secret) < 32 {
		log.Fatal("AUTH_SECRET is required and must be at least 32 characters")
	}

	ctx := context.Background()
	st, closeStore, err := openPersistence(ctx)
	if err != nil {
		log.Fatalf("open persistence: %v", err)
	}
	defer closeStore()

	hub := realtime.NewHub()
	go hub.Run()

	auth := service.NewAuthService(st, secret, 15*time.Minute, 7*24*time.Hour)
	svc := service.NewAgentWatchService(st, hub)

	if os.Getenv("DEMO_MODE") == "true" {
		if err := service.SeedDemo(auth, svc); err != nil {
			log.Fatalf("seed demo: %v", err)
		}
		log.Print("DEMO_MODE enabled: seeded demo account and sample data")
	}

	addr := env("ADDR", ":8080")
	server := httpapi.NewServer(auth, svc, hub)
	log.Printf("agentwatch api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.Handler()))
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
