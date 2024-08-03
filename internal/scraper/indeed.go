package scraper

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

type IndeedScraper struct{}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
}

func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

func (s *IndeedScraper) Scrape(config ScrapeConfig) ([]Job, error) {
	log.Printf("Starting Indeed scraper for job title: %s, country: %s, pages: %d", config.JobTitle, config.Country, config.Pages)

	c := colly.NewCollector(
		colly.UserAgent(getRandomUserAgent()),
		colly.AllowedDomains(fmt.Sprintf("%s.indeed.com", config.Country)),
	)

	// Rotate user agents
	extensions.RandomUserAgent(c)

	// Set custom headers
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
		r.Headers.Set("DNT", "1")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Cache-Control", "max-age=0")
	})

	transport := &http.Transport{
		DisableCompression: false,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	c.WithTransport(transport)

	c.SetRequestTimeout(60 * time.Second)

	c.Limit(&colly.LimitRule{
		RandomDelay: 5 * time.Second,
	})

	var jobs []Job

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Scraping: %s", r.URL)
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		log.Printf("Found job cards: %d", e.DOM.Find("#mosaic-provider-jobcards .job_seen_beacon").Length())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %s", r.Request.URL, err)
	})

	// Set up the scraping logic
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
