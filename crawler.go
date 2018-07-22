package scrape

import (
	"fmt"
	"log"

	"github.com/gocolly/colly"
)

// Crawler wrapper around colly
type Crawler struct {
	Collector *colly.Collector
	URL       string
	Debug     bool
	CacheDir  string
	Allowed   string
	Stats     int
}

func (c *Crawler) init() {

	c.Collector = colly.NewCollector(
		colly.CacheDir(c.CacheDir),
		colly.AllowedDomains(c.Allowed),
	)

	if c.Debug {
		c.Collector.OnRequest(func(r *colly.Request) {
			log.Printf("REQUEST: %v", r.URL.String())
		})

		c.Collector.OnResponse(func(r *colly.Response) {
			log.Printf("Response: %v", r.StatusCode)
		})

		c.Collector.OnError(func(r *colly.Response, err error) {
			log.Printf(`
		URL: 		  %v
		RESPONSE: %#v
		Error:    %v`,
				r.Request.URL, r, err)
		})
	}
}

func (c *Crawler) crawl() {
	// click on the video link
	c.Collector.OnHTML(".item-video .video-link", func(e *colly.HTMLElement) {
		url := e.Request.AbsoluteURL(e.Attr("href"))
		c.Collector.Visit(url)
	})
	// grap video info
	c.Collector.OnHTML("section.video-details", func(e *colly.HTMLElement) {
		fmt.Println(e.ChildText(".details > h1"))
	})
}

// Start starts the crawler
func (c *Crawler) Start() {
	c.init()
	c.crawl()
	c.Collector.Visit(c.URL)
}
