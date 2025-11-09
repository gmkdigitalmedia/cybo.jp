package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cyto-viewer/internal/api"
	"cyto-viewer/internal/config"
	"cyto-viewer/internal/scanner"
	"cyto-viewer/internal/tiler"
	"cyto-viewer/pkg/auth"
	"cyto-viewer/pkg/logger"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize logger
	log := logger.New()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize GPU tile processor
	tileProcessor, err := tiler.NewGPUTileProcessor(cfg.GPU)
	if err != nil {
		log.Fatal("Failed to initialize GPU processor", "error", err)
	}
	defer tileProcessor.Close()

	// Initialize scanner interface
	scannerInterface, err := scanner.NewInterface(cfg.Scanner)
	if err != nil {
		log.Fatal("Failed to initialize scanner interface", "error", err)
	}
	defer scannerInterface.Close()

	// Initialize authentication
	authManager := auth.NewManager(cfg.Auth)

	// Setup router
	router := mux.NewRouter()

	// API handlers
	apiHandler := api.NewHandler(log, tileProcessor, scannerInterface, authManager, cfg)
	apiHandler.RegisterRoutes(router)

	// Static files for the viewer
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", 
		http.FileServer(http.Dir("./web/static"))))

	// Main viewer page
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Starting server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Info("Server stopped")
}
