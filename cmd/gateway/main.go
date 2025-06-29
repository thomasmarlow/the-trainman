package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thomasmarlow/the-trainman/internal/config"
	"github.com/thomasmarlow/the-trainman/internal/server"
)

func main() {
	// initialize config manager
	configManager, err := config.NewManager("config.yaml")
	if err != nil {
		log.Fatalf("failed to create config manager: %v", err)
	}

	// start config watching
	if err := configManager.StartWatching(); err != nil {
		log.Fatalf("failed to start config watching: %v", err)
	}

	srv := server.NewServer(configManager)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: srv,
	}

	// channel to capture system signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// initiate server in goroutine
	go func() {
		log.Println("starting server on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// await shutdown signal
	<-quit
	log.Println("shutting down server...")

	// graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// stop config manager
	if err := configManager.Stop(); err != nil {
		log.Printf("error stopping config manager: %v", err)
	}

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
