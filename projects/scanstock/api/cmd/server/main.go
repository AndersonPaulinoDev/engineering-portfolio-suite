package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/AndersonPaulinoDev/scanstock/api/internal/auth"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/httpapi"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/service"
)

func main() {
	secret := strings.TrimSpace(os.Getenv("AUTH_SECRET"))
	if len(secret) < 32 {
		log.Fatal("AUTH_SECRET is required and must be at least 32 characters")
	}

	st, closeStore, err := openPersistence(context.Background())
	if err != nil {
		log.Fatalf("open persistence: %v", err)
	}
	defer closeStore()

	authService := auth.New(secret, st)
	inventoryService := service.New(st)
	if os.Getenv("DEMO_MODE") == "true" {
		if err := authService.SeedDemo(context.Background()); err != nil {
			log.Fatalf("seed demo users: %v", err)
		}
		if err := inventoryService.SeedDemo(context.Background()); err != nil {
			log.Fatalf("seed demo: %v", err)
		}
		log.Print("DEMO_MODE enabled: demo credentials are active")
	} else if email := strings.TrimSpace(os.Getenv("BOOTSTRAP_EMAIL")); email != "" {
		password := os.Getenv("BOOTSTRAP_PASSWORD")
		if len(password) < 8 {
			log.Fatal("BOOTSTRAP_PASSWORD must be at least 8 characters")
		}
		role := strings.TrimSpace(os.Getenv("BOOTSTRAP_ROLE"))
		if role == "" {
			role = "manager"
		}
		if err := authService.AddUser(context.Background(), "bootstrap_manager", email, password, role); err != nil {
			log.Fatalf("bootstrap user: %v", err)
		}
	}

	addr := strings.TrimSpace(os.Getenv("ADDR"))
	if addr == "" {
		addr = ":8082"
	}
	log.Printf("scanstock api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, httpapi.New(authService, inventoryService).Handler()))
}
