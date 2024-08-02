package storage

import (
	"fmt"
	"log"
	"sync"

	"github.com/ayagmar/gojobscraper/internal/scraper"
)

type MemoryStorage struct {
	jobs []scraper.Job
	mu   sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	log.Println("Initializing MemoryStorage")
	return &MemoryStorage{
		jobs: make([]scraper.Job, 0),
	}
}

func (m *MemoryStorage) SaveJobs(jobs []scraper.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobs = jobs // Replace existing jobs with new ones

	log.Printf("Saved %d new jobs in storage", len(jobs))
	return nil
}

func (m *MemoryStorage) GetJobs() ([]scraper.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jobCount := len(m.jobs)
	if jobCount == 0 {
		log.Println("No jobs found in storage")
		return nil, fmt.Errorf("no jobs found")
	}

	log.Printf("Retrieving %d jobs from storage", jobCount)
	return m.jobs, nil
}

func (m *MemoryStorage) ClearJobs() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobs = make([]scraper.Job, 0)
	log.Println("Cleared all jobs from storage")
	return nil
}
