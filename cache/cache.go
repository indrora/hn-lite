package cache

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"log"

	"github.com/cespare/xxhash/v2"
	"github.com/elastic/go-freelru"
	"github.com/gofiber/fiber/v2"
)

// CacheObject is the object that is stored in the cache
// It contains the content, the original URL, and the source URL
// OriginalUrl is the url of the content that is being cached
// SourceUrl is the url that requested the content to be cached
// CacheKey is the key that is used to store the content in the cache
// Stub is a boolean that indicates if the content is a stub or not
//
// Stub indicates that the content is empty and needs to be fetched from the original URL
type CacheObject struct {
	Content     []byte
	ContentType string
	Stub        bool
	ObjectUrl   string
	Referer     string
	CacheKey    string
}

func HashStringxx(s string) uint32 {
	return uint32(xxhash.Sum64String(s))
}

var lru *freelru.LRU[string, CacheObject]

const (
	CacheLifetime = 24 * time.Hour
	CacheSize     = 1024
)

func init() {
	var err error
	lru, err = freelru.New[string, CacheObject](CacheSize, HashStringxx)
	if err != nil {
		panic(err)
	}
}

func GetCacheObject(c *fiber.Ctx) error {
	hash := c.Params("hash")
	if obj, in := lru.GetAndRefresh(hash, CacheLifetime); in {

		// make sure it's not a stub
		if obj.Stub {
			// get the content from the original URL
			// using the referer header from the SourceUrl

			req, err := http.NewRequest("GET", obj.ObjectUrl, nil)

			if err != nil {
				return c.SendStatus(500)
			}
			if obj.Referer != "" {
				req.Header.Set("Referer", obj.Referer)
			}

			resp, err := http.DefaultClient.Do(req)

			if err != nil {

				return c.SendStatus(500)

			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return c.SendStatus(500)
			}
			obj.Stub = false
			obj.Content = body
			obj.ContentType = resp.Header.Get("content-type")
			lru.AddWithLifetime(hash, obj, CacheLifetime)

			c.Set("Content-Type", obj.ContentType)
			return c.Send(obj.Content)
		}

		c.Set("Cache-Control", "public, max-age=86400")
		c.Set("Content-Type", obj.ContentType)
		c.Set("Content-Length", string(len(obj.Content)))
		return c.Send(obj.Content)
	} else {
		return c.SendStatus(404)
	}
}

func AddCacheStub(source, referer url.URL) string {

	// hash the URL to get a unique key
	hash := sha256.Sum224([]byte(source.String()))
	hashstr := fmt.Sprintf("%x", hash)

	log.Printf("xxx %s", hashstr)
	log.Printf("from %v", source)
	log.Printf("referer %v", referer)
	// quickly check if we've already done this before...
	if _, in := lru.Peek(hashstr); in {
		log.Printf("Already cached %s", source.String())
		return hashstr
	}

	// otherwise, leave a stub in the cache

	obj := CacheObject{
		Stub:      true,
		ObjectUrl: source.String(),
		Referer:   referer.String(),
		CacheKey:  hashstr,
	}

	log.Printf("Caching stub for %s", source.String())
	lru.AddWithLifetime(hashstr, obj, CacheLifetime)

	return hashstr

}
