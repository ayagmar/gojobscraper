package scraper

import (
	"time"
)

type JobPosting struct {
	ID             string         `json:"id"`
	PlatformJobId  string         `json:"platform_job_id"`
	Title          string         `json:"title"`
	Company        string         `json:"company"`
	Location       string         `json:"location"`
	Summary        string         `json:"summary"`
	Description    string         `json:"description"`
	URL            string         `json:"url"`
	CompanyDetails CompanyDetails `json:"company_details"`
	Source         ScraperType    `json:"source"`
	CreatedAt      time.Time      `json:"createdAt"`
}
type CompanyDetails struct {
	PlatformCompanyURL string `json:"platform_company_url"`
	CompanyURL         string `json:"company_url"`
	CompanyIndustry    string `json:"company_industry"`
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
