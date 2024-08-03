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
// @Tags jobScraper
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
// @Tags jobScraper
// @Produce json
// @Param jobTitle query string true "Job Title"
// @Param country query string true "Country"
// @Param pages query int false "Number of Pages" default(1)
// @Param source query string true "Source of job listings (indeed or linkedin)" Enums(indeed, linkedin)
// @Success 202 {string} string "Scraping started"
// @Failure 400 {string} string "Bad Request"
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
	log.Printf("Starting scraping for job title: %s, country: %s, pages: %d, source: %s",
		config.JobTitle, config.Country, config.Pages, config.Source)

	scraper, err := scraper.NewScraper(config.Source)
	if err != nil {
		log.Printf("Error creating scraper: %v", err)
		return
	}

	jobs, err := scraper.Scrape(config)
	if err != nil {
		log.Printf("Error scraping %s: %v", config.Source, err)
		return
	}

	if err := h.storage.SaveJobs(jobs); err != nil {
		log.Printf("Error saving jobs: %v", err)
		return
	}

	log.Printf("Successfully scraped %d jobs from %s", len(jobs), config.Source)
}

func parseScrapingConfig(r *http.Request) (scraper.ScrapeConfig, error) {
	jobTitle := r.URL.Query().Get("jobTitle")
	country := r.URL.Query().Get("country")
	pagesStr := r.URL.Query().Get("pages")
	source := r.URL.Query().Get("source")

	if jobTitle == "" || country == "" || source == "" {
		return scraper.ScrapeConfig{}, errors.New("missing job title, country, or source")
	}

	pages, err := strconv.Atoi(pagesStr)
	if err != nil || pages < 1 {
		pages = 1 // Default to 1 page if not specified or invalid
	}

	scraperType := scraper.ScraperType(source)
	if scraperType != scraper.Indeed && scraperType != scraper.LinkedIn {
		return scraper.ScrapeConfig{}, errors.New("invalid source. Must be 'indeed' or 'linkedin'")
	}

	return scraper.ScrapeConfig{
		JobTitle: jobTitle,
		Country:  country,
		Pages:    pages,
		Source:   scraperType,
	}, nil
}

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response to JSON: %v", err)
	}
}
