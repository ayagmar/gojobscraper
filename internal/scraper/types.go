package scraper

import "time"

type Job struct {
	ID            string    `json:"id"`
	PlatformJobId string    `json:"platform_job_id"`
	Title         string    `json:"title"`
	Company       string    `json:"company"`
	Location      string    `json:"location"`
	Summary       string    `json:"summary"`
	URL           string    `json:"url"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ScrapeConfig struct {
	JobTitle string `json:"jobTitle"`
	Country  string `json:"country"`
	Pages    int    `json:"pages"`
}
