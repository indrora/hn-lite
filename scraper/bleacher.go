package scraper

import (
	"fmt"
	"log"
	"net/url"

	"github.com/indrora/hn-lite/cache"
	"github.com/microcosm-cc/bluemonday"
)

// Bleaching the HTML content

func Bleach(content string, originUrl url.URL) string {

	bluePolicy := bluemonday.UGCPolicy()
	bluePolicy.AllowTables()
	bluePolicy.AllowImages()

	bluePolicy.AllowElements("figure", "picture", "source")
	bluePolicy.AllowAttrs("srcset", "src", "type", "media").OnElements("source")
	bluePolicy.AllowNoAttrs().OnElements("div")

	bluePolicy.AllowRelativeURLs(true)

	bluePolicy.RequireNoFollowOnLinks(true)
	bluePolicy.RequireNoReferrerOnLinks(true)
	//bluePolicy.AllowRelativeURLs(false)
	bluePolicy.RequireParseableURLs(true)

	bluePolicy.RewriteSrc(func(u *url.URL) {

		log.Printf("Stubbing cache object for %v", u)
		cacheKey := cache.AddCacheStub(*u, originUrl)

		newUrl, err := url.Parse(fmt.Sprintf("/cache/%s", cacheKey))
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("xxxx %v", newUrl)
		*u = *newUrl

	})

	return bluePolicy.Sanitize(content)
}
