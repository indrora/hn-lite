package scraper

import (
	"log"
	"net/url"
	"time"

	"github.com/elastic/go-freelru"
	"github.com/go-shiori/go-readability"
	"github.com/gofiber/fiber/v2"
	"github.com/indrora/hn-lite/cache"
)

type ScrapedPage struct {
	Url        url.URL
	Content    string
	Title      string
	Author     string
	SiteName   string
	Date       *time.Time
	ScrapeTime time.Time
}

var pageLRU *freelru.LRU[string, ScrapedPage]

func init() {
	var err error
	pageLRU, err = freelru.New[string, ScrapedPage](1024, cache.HashStringxx)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			pageLRU.PurgeExpired()
			log.Printf("LRU size: %d", pageLRU.Len())
			pageLRU.PrintStats()
		}

	}()

}
func Scrape(u url.URL) (ScrapedPage, error) {
	// Get the URL from the request

	// check if it's already in the LRU

	if cached, in := pageLRU.GetAndRefresh(u.String(), 10*time.Minute); in {
		log.Printf("Cache hit for %s", u.String())
		return cached, nil
	}
	log.Printf("Cache miss for %s", u.String())

	// Get the page
	page, err := readability.FromURL(u.String(), 5*time.Second)
	if err != nil {
		log.Printf("Error getting page: %s", err)
		return ScrapedPage{}, err
	}

	// Bleach the HTML content
	bleached := Bleach(page.Content, u)

	scraped := ScrapedPage{
		Url:        u,
		Content:    bleached,
		Title:      page.Title,
		Author:     page.Byline,
		Date:       page.PublishedTime,
		SiteName:   page.SiteName,
		ScrapeTime: time.Now(),
	}

	// Add the page to the LRU
	pageLRU.AddWithLifetime(u.String(), scraped, 10*time.Minute)

	return scraped, nil
}

func RenderPage(c *fiber.Ctx) error {
	// Get the URL from the request
	urls := c.Query("url")

	url, err := url.Parse(urls)
	if err != nil {
		return c.SendStatus(400)
	}

	page, err := Scrape(*url)

	if err == nil {
		// Awesome
		return c.Render("proxy", fiber.Map{
			"title":     page.Title,
			"content":   page.Content,
			"url":       url,
			"author":    page.Author,
			"date":      page.Date,
			"scrapedon": page.ScrapeTime,
			"site":      page.SiteName,
		})
	} else {
		// Oh no
		return c.SendStatus(500)
	}
}
