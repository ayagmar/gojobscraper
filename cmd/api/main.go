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

	_ "github.com/ayagmar/gojobscraper/docs"
	"github.com/ayagmar/gojobscraper/internal/api"
	"github.com/ayagmar/gojobscraper/internal/config"
	"github.com/ayagmar/gojobscraper/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Job Scraper API
// @version 1.0
// @description This is a job scraper application.
// @termsOfService http://swagger.io/terms/
// @BasePath /api/v1
func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	log.Println("Starting Job Scraper Application")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	jobStorage, err := storage.NewMongoDBStorage(cfg.Database.URL, cfg.Database.Name)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer func(jobStorage *storage.MongoDBStorage) {
		err := jobStorage.Close()
		if err != nil {
			log.Fatalf("Error closing db storage: %v", err)
		}
	}(jobStorage)

	logger := log.New(os.Stdout, "JobScraper: ", log.LstdFlags|log.Lshortfile)

	router := setupRouter(jobStorage, logger)

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go startServer(srv, logger)
	waitForShutdown(srv, logger)

	return nil
}

func setupRouter(jobStorage storage.JobStorage, logger *log.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	handler := api.NewHandler(jobStorage, logger)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/jobs", handler.GetJobs)
		r.Post("/scrape", handler.StartScraping)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return r
}

func startServer(srv *http.Server, logger *log.Logger) {
	logger.Printf("Starting HTTP server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Error starting server: %v", err)
	}
}

func waitForShutdown(srv *http.Server, logger *log.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Println("Shutting down the application...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Error during server shutdown: %v", err)
	}

	logger.Println("Application stopped")
}
