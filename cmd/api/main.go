package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayagmar/gojobscraper/internal/api"
	"github.com/ayagmar/gojobscraper/internal/storage"
)

func main() {
	log.Println("Starting Job Scraper Application")

	// Initialize storage
	connStr := "postgres://jobuser:jobpassword@localhost:5432/jobscraper?sslmode=disable"
	jobStorage, err := storage.NewPostgresStorage(connStr)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize API handler
	handler := &api.Handler{
		Storage: jobStorage,
	}

	// Set up HTTP routes
	http.HandleFunc("/jobs", handler.GetJobs)
	http.HandleFunc("/scrape", handler.StartScraping)

	// Create a channel to signal application shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down the application...")

	// Gracefully shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Application stopped")
}
