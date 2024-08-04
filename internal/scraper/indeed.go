package scraper

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type IndeedScraper struct{}

func (s *IndeedScraper) Scrape(config ScrapeConfig) ([]JobPosting, error) {
	log.Printf("Starting Indeed scraper for job title: %s, country: %s, pages: %d", config.JobTitle, config.Country, config.Pages)

	c := SetupColly(fmt.Sprintf("%s.indeed.com", config.Country))
	if c == nil {
		return nil, fmt.Errorf("failed to setup collector")
	}

	jobs := make([]JobPosting, 0)

	c.OnHTML("#mosaic-provider-jobcards .job_seen_beacon", func(e *colly.HTMLElement) {
		job, err := s.parseJobCard(e)
		if err != nil {
			log.Printf("Error parsing job card: %v", err)
			return
		}
		jobs = append(jobs, job)
		log.Printf("Parsed job: %s at %s, URL: %s", job.Title, job.Company, job.URL)
	})

	err := s.visitPages(c, config)
	if err != nil {
		return nil, fmt.Errorf("error visiting pages: %w", err)
	}

	log.Printf("Scraped total of %d jobs from Indeed", len(jobs))
	return jobs, nil
}

func (s *IndeedScraper) parseJobCard(e *colly.HTMLElement) (JobPosting, error) {
	dirtyURL := e.Request.AbsoluteURL(e.ChildAttr("h2.jobTitle a", "href"))
	cleanURL := cleanJobURL(dirtyURL)

	job := JobPosting{
		ID:            uuid.New().String(),
		PlatformJobId: ExtractJobKey(cleanURL),
		Title:         e.ChildText(".jobTitle span"),
		Company:       e.ChildText("[data-testid='company-name']"),
		Location:      e.ChildText("[data-testid='text-location']"),
		Summary:       e.ChildText(".css-9446fg"),
		URL:           cleanURL,
		CreatedAt:     time.Now(),
		Source:        Indeed,
	}

	description, companyDetails, err := s.fetchJobDetails(job.URL)
	if err != nil {
		return JobPosting{}, fmt.Errorf("error fetching job description: %w", err)
	}

	job.Description = description
	job.CompanyDetails = companyDetails

	return job, nil
}

func (s *IndeedScraper) fetchJobDetails(jobURL string) (string, CompanyDetails, error) {
	c := SetupColly("www.indeed.com")
	if c == nil {
		return "", CompanyDetails{}, fmt.Errorf("failed to setup collector for job description")
	}

	var description string
	var companyDetails CompanyDetails

	c.OnHTML("#jobDescriptionText", func(e *colly.HTMLElement) {
		description = strings.TrimSpace(e.Text)
	})

	c.OnHTML("div[data-company-name='true']", func(e *colly.HTMLElement) {
		dirtyCompanyURL := e.ChildAttr("a", "href")
		companyDetails.PlatformCompanyURL = cleanCompanyURL(dirtyCompanyURL)
	})

	err := c.Visit(jobURL)
	if err != nil {
		return "", CompanyDetails{}, err
	}

	err = s.fetchCompanyDetails(&companyDetails)
	if err != nil {
		log.Printf("Error fetching company details: %v", err)
	}

	return description, companyDetails, nil
}
func (s *IndeedScraper) visitPages(c *colly.Collector, config ScrapeConfig) error {
	baseURL := fmt.Sprintf("https://%s.indeed.com/jobs", config.Country)
	query := url.Values{}
	query.Set("q", config.JobTitle)

	for page := 0; page < config.Pages; page++ {
		query.Set("start", fmt.Sprintf("%d", page*10))
		fullURL := fmt.Sprintf("%s?%s", baseURL, query.Encode())

		err := c.Visit(fullURL)
		if err != nil {
			log.Printf("Error visiting page %d: %v", page, err)
			if page == 0 {
				return fmt.Errorf("error visiting first page: %w", err)
			}
			continue
		}

		// Add a longer delay between pages
		time.Sleep(time.Duration(rand.Intn(5)+5) * time.Second)
	}

	return nil
}

func (s *IndeedScraper) fetchCompanyDetails(details *CompanyDetails) error {
	c := SetupColly("www.indeed.com")
	if c == nil {
		return fmt.Errorf("failed to setup collector for company details")
	}

	c.OnHTML("li[data-testid='companyInfo-industry']", func(e *colly.HTMLElement) {
		details.CompanyIndustry = e.ChildText("div.css-kaq73 a")
	})

	c.OnHTML("li[data-testid='companyInfo-companyWebsite']", func(e *colly.HTMLElement) {
		details.CompanyURL = e.ChildAttr("div.css-kaq73 a", "href")
	})

	err := c.Visit(details.PlatformCompanyURL)
	if err != nil {
		return fmt.Errorf("error visiting company page: %w", err)
	}

	return nil
}

func cleanJobURL(dirtyURL string) string {
	parsedURL, err := url.Parse(dirtyURL)
	if err != nil {
		log.Printf("Error parsing URL: %s", err)
		return dirtyURL
	}

	queryParams, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		log.Printf("Error parsing query parameters: %s", err)
		return dirtyURL
	}

	jk := queryParams.Get("jk")
	if jk == "" {
		log.Printf("No 'jk' parameter found in URL: %s", dirtyURL)
		return dirtyURL
	}

	return fmt.Sprintf("https://www.indeed.com/viewjob?jk=%s", jk)
}

func ExtractJobKey(jobUrl string) string {
	parsedURL, err := url.Parse(jobUrl)
	if err != nil {
		log.Printf("Error parsing URL: %s", err)
		return ""
	}

	queryParams, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		log.Printf("Error parsing query parameters: %s", err)
		return ""
	}

	jk := queryParams.Get("jk")
	if jk == "" {
		log.Printf("No 'jk' parameter found in URL: %s", jobUrl)
		return ""
	}

	return jk
}

func cleanCompanyURL(dirtyURL string) string {
	parsedURL, err := url.Parse(dirtyURL)
	if err != nil {
		log.Printf("Error parsing company URL: %s", err)
		return dirtyURL
	}

	// Keep only the scheme and path
	cleanedURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	// Remove trailing slash if present
	return strings.TrimSuffix(cleanedURL, "/")
}
