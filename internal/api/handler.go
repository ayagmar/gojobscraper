package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/ayagmar/gojobscraper/internal/scraper"
	"github.com/ayagmar/gojobscraper/internal/storage"
	"github.com/go-chi/render"
)

// Handler manages HTTP requests for the job scraper API.
type Handler struct {
	storage storage.JobStorage
	logger  *log.Logger
}

// NewHandler creates a new Handler instance.
func NewHandler(storage storage.JobStorage, logger *log.Logger) *Handler {
	return &Handler{storage: storage, logger: logger}
}

// GetJobs handles GET requests for retrieving jobs.
// @Summary Get jobs
// @Description Get a list of jobs
// @Tags jobScraper
// @Accept json
// @Produce json
// @Success 200 {array} scraper.Job
// @Failure 500 {object} ErrorResponse
// @Router /jobs [get]
func (h *Handler) GetJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.storage.GetJobs()
	if err != nil {
		h.logger.Printf("Error retrieving jobs: %v", err)
		render.Render(w, r, ErrInternalServer(err))
		return
	}

	render.JSON(w, r, jobs)
}

// StartScraping handles POST requests to initiate job scraping.
// @Summary Start scraping
// @Description Start scraping jobs based on the provided configuration
// @Tags jobScraper
// @Accept json
// @Produce json
// @Param jobTitle query string true "Job Title"
// @Param country query string true "Country"
// @Param pages query int false "Number of Pages" default(1)
// @Param source query string true "Source of job listings (indeed or linkedin)" Enums(indeed, linkedin)
// @Success 202 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 501 {object} ErrorResponse
// @Router /scrape [post]
func (h *Handler) StartScraping(w http.ResponseWriter, r *http.Request) {
	config, err := h.parseScrapingConfig(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	go h.scrapeAndSaveJobs(config)

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, SuccessResponse{Message: "Scraping started"})
}

func (h *Handler) scrapeAndSaveJobs(config scraper.ScrapeConfig) {
	h.logger.Printf("Starting scraping for job title: %s, country: %s, pages: %d, source: %s",
		config.JobTitle, config.Country, config.Pages, config.Source)

	scraper, err := scraper.NewScraper(config.Source)
	if err != nil {
		h.logger.Printf("Error creating scraper: %v", err)
		return
	}

	jobs, err := scraper.Scrape(config)
	if err != nil {
		h.logger.Printf("Error scraping %s: %v", config.Source, err)
		return
	}

	if err := h.storage.SaveJobs(jobs); err != nil {
		h.logger.Printf("Error saving jobs: %v", err)
		return
	}

	h.logger.Printf("Successfully scraped %d jobs from %s", len(jobs), config.Source)
}

func (h *Handler) parseScrapingConfig(r *http.Request) (scraper.ScrapeConfig, error) {
	query := r.URL.Query()
	jobTitle := query.Get("jobTitle")
	country := query.Get("country")
	pagesStr := query.Get("pages")
	source := query.Get("source")

	if jobTitle == "" || country == "" || source == "" {
		return scraper.ScrapeConfig{}, errors.New("missing required parameters: job title, country, or source")
	}

	pages, err := strconv.Atoi(pagesStr)
	if err != nil || pages < 1 {
		pages = 1 // Default to 1 page if not specified or invalid
	}

	scraperType := scraper.ScraperType(source)
	if !isValidScraperType(scraperType) {
		return scraper.ScrapeConfig{}, errors.New("invalid source. Must be 'indeed' or 'linkedin'")
	}

	return scraper.ScrapeConfig{
		JobTitle: jobTitle,
		Country:  country,
		Pages:    pages,
		Source:   scraperType,
	}, nil
}

func isValidScraperType(t scraper.ScraperType) bool {
	return t == scraper.Indeed || t == scraper.LinkedIn
}

type ErrorResponse struct {
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status"`
	ErrorText      string `json:"error,omitempty"`
}

func (e *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrorResponse{
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Invalid request",
		ErrorText:      err.Error(),
	}
}

func ErrInternalServer(err error) render.Renderer {
	return &ErrorResponse{
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "Internal server error",
		ErrorText:      err.Error(),
	}
}

func ErrNotImplemented(err error) render.Renderer {
	return &ErrorResponse{
		HTTPStatusCode: http.StatusNotImplemented,
		StatusText:     "Not Implemented",
		ErrorText:      err.Error(),
	}
}

type SuccessResponse struct {
	Message string `json:"message"`
}
