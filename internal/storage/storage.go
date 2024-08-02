package storage

import (
	"github.com/ayagmar/gojobscraper/internal/scraper"
)

type JobStorage interface {
	SaveJobs(jobs []scraper.Job) error
	GetJobs() ([]scraper.Job, error)
	ClearJobs() error
}
