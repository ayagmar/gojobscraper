package scraper

import "time"

type Job struct {
	ID            string      `json:"id"`
	PlatformJobId string      `json:"platform_job_id"`
	Title         string      `json:"title"`
	Company       string      `json:"company"`
	Location      string      `json:"location"`
	Summary       string      `json:"summary"`
	URL           string      `json:"url"`
	Source        ScraperType `json:"source"`
	CreatedAt     time.Time   `json:"createdAt"`
}

type ScraperType string

const (
	Indeed   ScraperType = "indeed"
	LinkedIn ScraperType = "linkedin"
)

type ScrapeConfig struct {
	JobTitle string
	Country  string
	Pages    int
	Source   ScraperType
}

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
