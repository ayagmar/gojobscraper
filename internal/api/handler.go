package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/ayagmar/gojobscraper/internal/scraper"
	"github.com/ayagmar/gojobscraper/internal/storage"
)

type Handler struct {
	storage storage.JobStorage
}

func NewHandler(storage storage.JobStorage) *Handler {
	return &Handler{storage: storage}
}

// GetJobs godoc
// @Summary Get jobs
// @Description Get a list of jobs
// @Tags jobs
// @Accept json
// @Produce json
// @Success 200 {array} scraper.Job
// @Router /jobs [get]
func (h *Handler) GetJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.storage.GetJobs()
	if err != nil {
		log.Printf("Error retrieving jobs: %v", err)
		http.Error(w, "Failed to retrieve jobs", http.StatusInternalServerError)
		return
	}

	respondJSON(w, jobs, http.StatusOK)
}

// StartScraping godoc
// @Summary Start scraping
// @Description Start scraping jobs based on the provided configuration
// @Tags scrape
// @Accept json
// @Produce json
// @Param jobTitle query string true "Job Title"
// @Param country query string true "Country"
// @Param pages query int false "Number of Pages"
// @Success 202 {string} string "Scraping started"
// @Router /scrape [post]
func (h *Handler) StartScraping(w http.ResponseWriter, r *http.Request) {
	config, err := parseScrapingConfig(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go h.scrapeAndSaveJobs(config)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Scraping started"))
}

func (h *Handler) scrapeAndSaveJobs(config scraper.ScrapeConfig) {
	log.Printf("Starting scraping for job title: %s, country: %s, pages: %d", config.JobTitle, config.Country, config.Pages)
	jobs, err := scraper.ScrapeIndeed(config)
	if err != nil {
		log.Printf("Error scraping Indeed: %v", err)
		return
	}

	if err := h.storage.SaveJobs(jobs); err != nil {
		log.Printf("Error saving jobs: %v", err)
		return
	}

	log.Printf("Successfully scraped %d jobs", len(jobs))
}

func parseScrapingConfig(r *http.Request) (scraper.ScrapeConfig, error) {
	jobTitle := r.URL.Query().Get("title")
	country := r.URL.Query().Get("country")
	pagesStr := r.URL.Query().Get("pages")

	if jobTitle == "" || country == "" {
		return scraper.ScrapeConfig{}, errors.New("missing job title or country")
	}

	pages, err := strconv.Atoi(pagesStr)
	if err != nil || pages < 1 {
		pages = 1 // Default to 1 page if not specified or invalid
	}

	return scraper.ScrapeConfig{
		JobTitle: jobTitle,
		Country:  country,
		Pages:    pages,
	}, nil
}

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response to JSON: %v", err)
	}
}
