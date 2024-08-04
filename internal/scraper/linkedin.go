package scraper

import (
	"errors"
)

type LinkedInScraper struct{}

func (s *LinkedInScraper) Scrape(config ScrapeConfig) ([]JobPosting, error) {
	return nil, errors.New("LinkedIn scraper not implemented yet")
}
