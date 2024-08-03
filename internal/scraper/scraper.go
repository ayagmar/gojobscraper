package scraper

import (
	"fmt"
)

type Scraper interface {
	Scrape(config ScrapeConfig) ([]Job, error)
}

func NewScraper(scraperType ScraperType) (Scraper, error) {
	switch scraperType {
	case Indeed:
		return &IndeedScraper{}, nil
	case LinkedIn:
		return &LinkedInScraper{}, nil
	default:
		return nil, fmt.Errorf("unsupported scraper type: %s", scraperType)
	}
}
