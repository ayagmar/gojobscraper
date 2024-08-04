package storage

import (
	"github.com/ayagmar/gojobscraper/internal/scraper"
)

type JobStorage interface {
	SaveJobs(jobs []scraper.JobPosting) error
	GetJobs() ([]scraper.JobPosting, error)
	ClearJobs() error
	Close() error
}
