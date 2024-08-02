package scraper

import "time"

type Job struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Company   string    `json:"company"`
	Location  string    `json:"location"`
	Salary    string    `json:"salary"`
	Summary   string    `json:"summary"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
}

type ScrapeConfig struct {
	JobTitle string `json:"jobTitle"`
	Country  string `json:"country"`
	Pages    int    `json:"pages"`
}
