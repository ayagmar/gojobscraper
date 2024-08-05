package scraper

import (
	"time"
)

// JobPosting represents a job posting
// @Description Job posting details
type JobPosting struct {
	ID             string         `json:"id"`
	PlatformJobId  string         `json:"platform_job_id"`
	Title          string         `json:"title"`
	Location       string         `json:"location"`
	Summary        string         `json:"summary"`
	Description    string         `json:"description"`
	URL            string         `json:"url"`
	CompanyDetails CompanyDetails `json:"company_details"`
	Source         ScraperType    `json:"source"`
	CreatedAt      time.Time      `json:"createdAt"`
}

// CompanyDetails represents details about a company
// @Description Company details
type CompanyDetails struct {
	PlatformCompanyURL string `json:"platform_company_url"`
	CompanyURL         string `json:"url"`
	CompanyIndustry    string `json:"industry"`
	Company            string `json:"name"`
}

// ScraperType represents the type of job scraper
// @Description Type of job scraper
type ScraperType string

const (
	Indeed   ScraperType = "indeed"
	LinkedIn ScraperType = "linkedin"
)

// ScrapeConfig represents the configuration for a job scraping operation
// @Description Configuration for job scraping
type ScrapeConfig struct {
	JobTitle string
	Country  string
	Pages    int
	Source   ScraperType
}
