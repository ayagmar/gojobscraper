package scraper

import (
	"time"
)

type JobPosting struct {
	ID            string      `json:"id"`
	PlatformJobId string      `json:"platform_job_id"`
	Title         string      `json:"title"`
	Company       string      `json:"company"`
	Location      string      `json:"location"`
	Summary       string      `json:"summary"`
	Description   string      `json:"description"`
	URL           string      `json:"url"`
	Source        ScraperType `json:"source"`
	CreatedAt     time.Time   `json:"createdAt"`
	ID                 string      `json:"id"`
	PlatformJobId      string      `json:"platform_job_id"`
	Title              string      `json:"title"`
	Company            string      `json:"company"`
	Location           string      `json:"location"`
	Summary            string      `json:"summary"`
	Description        string      `json:"description"`
	URL                string      `json:"url"`
	PlatformCompanyURL string      `json:"platform_company_url"`
	Source             ScraperType `json:"source"`
	CreatedAt          time.Time   `json:"createdAt"`
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
