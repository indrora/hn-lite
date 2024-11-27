package main

import (
	"log"

	"html/template"

	"github.com/gofiber/fiber/v2"
	fiberHtml "github.com/gofiber/template/html/v2"
	"github.com/indrora/hn-lite/cache"
	"github.com/indrora/hn-lite/scraper"
)

func main() {
	// Initialize a new Fiber app

	templateEngine := fiberHtml.New("./templates", ".html")

	// Initialize a new LRU cache with a capacity of 1000 items

	templateEngine.AddFunc("safe", func(s string) template.HTML {
		return template.HTML(s)
	})

	app := fiber.New(fiber.Config{
		Views: templateEngine,
	})

	// Define a route for the GET method on the root path '/'
	app.Get("/", func(c *fiber.Ctx) error {
		// Send a string response to the client
		return c.SendString("Hello, World 👋!")
	})

	// static content
	app.Static("/static", "./static")

	app.Get("/cache/:hash", cache.GetCacheObject)

	app.Get("/proxy", scraper.RenderPage)

	// Start the server on port 3000
	log.Fatal(app.Listen(":3000"))
}
