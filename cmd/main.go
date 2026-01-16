package main

import (
	"log"

	"go-backend/config"
	"go-backend/internal/db"
	"go-backend/internal/server"
	"go-backend/internal/service"
)

func main() {
	cfg := config.Load()

	store := db.NewDataStore()
	svc := service.New(store)
	srv := server.New(svc, cfg)

	if err := srv.Start(); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
