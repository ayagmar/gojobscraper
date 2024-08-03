package scraper

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
)

type IndeedScraper struct{}

func (s *IndeedScraper) Scrape(config ScrapeConfig) ([]Job, error) {
	log.Printf("Starting Indeed scraper for job title: %s, country: %s, pages: %d", config.JobTitle, config.Country, config.Pages)

	c := SetupColly(fmt.Sprintf("%s.indeed.com", config.Country))

	var jobs []Job

	c.OnHTML("body", func(e *colly.HTMLElement) {
		log.Printf("Found job cards: %d", e.DOM.Find("#mosaic-provider-jobcards .job_seen_beacon").Length())
	})

	c.OnHTML("#mosaic-provider-jobcards .job_seen_beacon", func(e *colly.HTMLElement) {
		dirtyURL := e.Request.AbsoluteURL(e.ChildAttr("h2.jobTitle a", "href"))
		cleanURL := cleanJobURL(dirtyURL)
		job := Job{
			PlatformJobId: ExtractJobKey(cleanURL),
			Title:         e.ChildText(".jobTitle span"),
			Company:       e.ChildText("[data-testid='company-name']"),
			Location:      e.ChildText("[data-testid='text-location']"),
			Summary:       e.ChildText(".css-9446fg"),
			URL:           cleanURL,
			CreatedAt:     time.Now(),
			Source:        Indeed,
		}
		jobs = append(jobs, job)
		log.Printf("Parsed job: %s at %s, URL: %s", job.Title, job.Company, job.URL)
	})

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
				return nil, err // If we can't scrape the first page, return the error
			}
			continue
		}

		// Add a longer delay between pages
		time.Sleep(time.Duration(rand.Intn(5)+5) * time.Second)
	}

	log.Printf("Scraped total of %d jobs from Indeed", len(jobs))
	return jobs, nil
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
