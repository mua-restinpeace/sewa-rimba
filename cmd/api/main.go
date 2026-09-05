package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/mua-restinpeace/sewa-rimba/internal/config"
	"github.com/mua-restinpeace/sewa-rimba/internal/handler"
	"github.com/mua-restinpeace/sewa-rimba/internal/repository"
	"github.com/mua-restinpeace/sewa-rimba/internal/router"
	"github.com/mua-restinpeace/sewa-rimba/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// repositories
	equipmentRepo := repository.NewEquipmentrRepository(db)

	// services
	equipmentService := service.NewEquipmentService(equipmentRepo)

	// handlers
	handlers := router.Handlers{
		Equipment: handler.NewEquipmentHandler(equipmentService),
	}

	r := router.New(handlers)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("listening on: %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shotdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shotdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
