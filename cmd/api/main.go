package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/ayagmar/gojobscraper/docs"
	"github.com/ayagmar/gojobscraper/internal/api"
	"github.com/ayagmar/gojobscraper/internal/config"
	"github.com/ayagmar/gojobscraper/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger" // Import http-swagger
)

// @title Job Scraper API
// @version 1.0
// @description This is a job scraper application.
// @termsOfService http://swagger.io/terms/
// @host localhost:8080
// @BasePath /api/v1
func main() {
	log.Println("Starting Job Scraper Application")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	jobStorage, err := storage.NewPostgresStorage(cfg.Database.URL)
	log.Println(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer jobStorage.Close()

	router := setupRouter(jobStorage)

	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	go startServer(srv)
	waitForShutdown(srv)
}

func setupRouter(jobStorage storage.JobStorage) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	handler := api.NewHandler(jobStorage)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/jobs", handler.GetJobs)
		r.Post("/scrape", handler.StartScraping)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The url pointing to API definition
	))

	return r
}

func startServer(srv *http.Server) {
	log.Printf("Starting HTTP server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting server: %v", err)
	}
}

func waitForShutdown(srv *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down the application...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Application stopped")
}
