package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/ayagmar/gojobscraper/internal/scraper"
	"github.com/ayagmar/gojobscraper/internal/storage"
)

type Handler struct {
	Storage storage.JobStorage
}

func (h *Handler) GetJobs(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request for jobs. Method: %s, URL: %s", r.Method, r.URL)

	if r.Method != http.MethodGet {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs, err := h.Storage.GetJobs()
	if err != nil {
		log.Printf("Error retrieving jobs: %v", err)
		http.Error(w, "Failed to retrieve jobs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jobs); err != nil {
		log.Printf("Error encoding jobs to JSON: %v", err)
		http.Error(w, "Failed to encode jobs", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully returned %d jobs", len(jobs))
}

func (h *Handler) StartScraping(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request to start scraping. Method: %s, URL: %s", r.Method, r.URL)

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobTitle := r.URL.Query().Get("title")
	country := r.URL.Query().Get("country")
	pagesStr := r.URL.Query().Get("pages")

	if jobTitle == "" || country == "" {
		log.Println("Missing job title or country")
		http.Error(w, "Missing job title or country", http.StatusBadRequest)
		return
	}

	pages, err := strconv.Atoi(pagesStr)
	if err != nil || pages < 1 {
		log.Println("Invalid number of pages")
		pages = 1 // Default to 1 page if not specified or invalid
	}

	config := scraper.ScrapeConfig{
		JobTitle: jobTitle,
		Country:  country,
		Pages:    pages,
	}

	go func() {
		log.Printf("Starting scraping for job title: %s, country: %s, pages: %d", config.JobTitle, config.Country, config.Pages)
		jobs, err := scraper.ScrapeIndeed(config)
		if err != nil {
			log.Printf("Error scraping Indeed: %v", err)
			return
		}

		/* 		if err := h.Storage.ClearJobs(); err != nil {
			log.Printf("Error clearing jobs: %v", err)
			return
		} */

		if err := h.Storage.SaveJobs(jobs); err != nil {
			log.Printf("Error saving jobs: %v", err)
			return
		}

		log.Printf("Successfully scraped and saved %d jobs", len(jobs))
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Scraping started"))
}
