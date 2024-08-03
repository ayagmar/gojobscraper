package scraper

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/corpix/uarand"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
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

func getRandomUserAgent() string {
	return uarand.GetRandom()
}

func SetupColly(allowedDomains ...string) *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent(getRandomUserAgent()),
		colly.AllowedDomains(allowedDomains...),
		colly.MaxDepth(2),
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

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Scraping: %s", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %s", r.Request.URL, err)
	})

	return c
}
