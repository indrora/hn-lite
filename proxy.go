package main

import (
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2/log"

	"embed"
	"html/template"

	"github.com/gofiber/fiber/v2"
	fiberHtml "github.com/gofiber/template/html/v2"
	"github.com/indrora/hn-lite/cache"
	"github.com/indrora/hn-lite/reload"
	"github.com/indrora/hn-lite/scraper"
	"github.com/indrora/hn-lite/stories"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {

	//templateEngine := fiberHtml.New("./templates", ".html")

	tfs, err := fs.Sub(templateFS, "templates")
	if err != nil {
		panic(err)
	}
	templateEngine := fiberHtml.NewFileSystem(http.FS(tfs), ".html")

	if err != nil {
		panic(err)
	}

	// Initialize a new LRU cache with a capacity of 1000 items

	templateEngine.AddFunc("safe", func(s string) template.HTML {
		return template.HTML(s)
	})

	templateEngine.AddFunc("cachepls", func(s string) template.HTML {
		// Preload the cache with this item
		uu, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		return template.HTML(cache.AddCacheStub(uu, nil))
	})

	app := fiber.New(fiber.Config{
		Views: templateEngine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		// Get top stories

		stories := stories.TopStories()
		log.Debugf("%v+", stories)

		return c.Render("index", fiber.Map{
			"stories": stories,
		}, "layout.main")
	})

	// Cache layer
	app.Get("/cache/:hash", cache.GetCacheObject)
	// Metrics that are interesting
	app.Get("/metrics", StatsPage)
	// Actual proxy
	app.Get("/proxy", scraper.RenderPage)
	// live reload hack...
	reload.Attach("/reload", app, templateEngine)

	app.Get("/*", func(c *fiber.Ctx) error {
		// Send a string response to the client

		// Get the URL from the request
		pg := c.Params("*1", "")
		// Check if the URL is valid
		if pg == "" {
			return c.SendStatus(404)
		}
		log.Infof("Unstructured request for %v", pg)
		if strings.HasPrefix(pg, "static") {
			// try to get the item out of the staticfs
			fr, err := staticFS.ReadFile(pg)
			if err != nil {
				c.SendStatus(404)
				log.Error(err)
			} else {
				return c.Send(fr)
			}
		}
		err := c.Render(pg, fiber.Map{}, "layout.main")
		if err != nil {
			log.Error(err)
			return c.SendStatus(500)
		} else {
			return err
		}
	})

	// Start the server on port 3000
	log.Fatal(app.Listen(":3000"))
}

func StatsPage(c *fiber.Ctx) error {
	// get the stats out of each of the caches

	mediametrics := cache.GetStats()
	pagemetrics := scraper.GetStats()

	mediaHitRatio := (mediametrics.Hits / max(mediametrics.Hits+mediametrics.Misses, 1)) * 100.0
	pageHitRatio := (pagemetrics.Hits / max(pagemetrics.Hits+pagemetrics.Misses, 1)) * 100.0

	return c.Render("metrics",
		fiber.Map{
			"media": fiber.Map{
				"hits":     mediametrics.Hits,
				"misses":   mediametrics.Misses,
				"adds":     mediametrics.Inserts,
				"evicts":   mediametrics.Evictions,
				"removals": mediametrics.Removals,
				"ratio":    mediaHitRatio,
			},
			"page": fiber.Map{
				"hits":     pagemetrics.Hits,
				"misses":   pagemetrics.Misses,
				"adds":     pagemetrics.Inserts,
				"evicts":   pagemetrics.Evictions,
				"removals": pagemetrics.Removals,
				"ratio":    pageHitRatio,
			},
		}, "layout.main")

}
