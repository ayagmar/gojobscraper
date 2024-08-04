package storage

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ayagmar/gojobscraper/internal/scraper"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")

	storage := &PostgresStorage{db: db}
	if err := storage.createJobsTable(); err != nil {
		return nil, fmt.Errorf("failed to create jobs table: %w", err)
	}

	return storage, nil
}

func (p *PostgresStorage) createJobsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			platform_job_id TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			company TEXT NOT NULL,
			location TEXT NOT NULL,
			summary TEXT,
			description TEXT,
			url TEXT NOT NULL,
			source TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`
	_, err := p.db.Exec(query)
	return err
}

func (p *PostgresStorage) SaveJobs(jobs []scraper.JobPosting) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
	INSERT INTO jobs (id, platform_job_id, title, company, location, summary, description, url, source, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT (platform_job_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, job := range jobs {
		if job.ID == "" {
			job.ID = uuid.New().String()
		}

		result, err := stmt.Exec(job.ID, job.PlatformJobId, job.Title, job.Company, job.Location, job.Summary, job.Description, job.URL, job.Source, job.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert job: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			log.Printf("JobPosting not saved due to conflict: PlatformJobId=%s, Title=%s", job.PlatformJobId, job.Title)
		} else {
			log.Printf("Saved job: ID=%s, Title=%s", job.ID, job.Title)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Processed %d jobs in PostgreSQL storage", len(jobs))
	return nil
}

func (p *PostgresStorage) GetJobs() ([]scraper.JobPosting, error) {
	query := `
		SELECT id, platform_job_id, title, company, location, summary, description, url, source, created_at
		FROM jobs
		ORDER BY created_at DESC
	`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []scraper.JobPosting
	for rows.Next() {
		var job scraper.JobPosting
		err := rows.Scan(&job.ID, &job.PlatformJobId, &job.Title, &job.Company, &job.Location, &job.Summary, &job.Description, &job.URL, &job.Source, &job.CreatedAt)
		if err != nil {
			log.Printf("Error scanning job row: %v", err)
			continue
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating job rows: %w", err)
	}

	jobCount := len(jobs)
	log.Printf("Retrieved %d jobs from PostgreSQL storage", jobCount)

	return jobs, nil
}

func (p *PostgresStorage) ClearJobs() error {
	_, err := p.db.Exec("DELETE FROM jobs")
	if err != nil {
		return fmt.Errorf("failed to clear jobs: %w", err)
	}

	log.Println("Cleared all jobs from PostgreSQL storage")
	return nil
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}
