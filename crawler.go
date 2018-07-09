package souqr

import (
	"fmt"
	"log"
	"time"

	tor "github.com/allyraza/souqr/proxy"
	"github.com/garyburd/redigo/redis"
	"github.com/gocolly/colly"
	//	"github.com/gocolly/redisstorage"
)

// Crawler wrapper around colly
type Crawler struct {
	URL         string
	Proxy       string
	Tor         string
	TorControl  string
	Collector   *colly.Collector
	Timeout     time.Duration
	Parallelism int
	RandomDelay time.Duration
	Delay       time.Duration
	CacheDir    string
	URLFilters  string
	Allowed     string
	UserAgent   string
	Store       redis.Conn
	RedisURL    string
	Debug       bool
	PerPage     int
	Stats       int
}

func (c *Crawler) init() {
	// rc, err := redis.Dial("tcp", c.RedisURL)
	// if err != nil {
	// 	log.Fatalf("Could not connect to redis: %v", err)
	// }

	// c.Store = rc
	// c.Stats = 0

	c.Collector = colly.NewCollector(
		// cache directory
		colly.CacheDir(c.CacheDir),
		// allowed domains
		colly.AllowedDomains(c.Allowed),
		// url filters
		// colly.URLFilters(regexp.MustCompile(c.URLFilters)),
		// user agent
		colly.UserAgent(c.UserAgent),
	)

	if c.Debug {
		c.Collector.OnRequest(func(r *colly.Request) {
			log.Printf("REQUEST: %v", r.URL.String())
		})
		c.Collector.OnResponse(func(r *colly.Response) {
			log.Printf("Response: %v", r.StatusCode)
		})
	}

	// storage := &redisstorage.Storage{
	// 	Address: c.RedisURL,
	// 	Prefix:  "souqr_",
	// }
	// defer storage.Client.Close()
	//
	// if err = c.Pc.SetStorage(storage); err != nil {
	// 	panic(err)
	// }
	//
	// if err := storage.Clear(); err != nil {
	// 	log.Fatal(err)
	// }

	c.Collector.OnError(func(r *colly.Response, err error) {
		log.Printf(`
		URL: 		  %v
		RESPONSE: %#v
		Error:    %v`,
			r.Request.URL, r, err)
	})
}

func (c *Crawler) setProxy() {
	if c.Tor != "" {
		pf, err := tor.TorProxySwitcher(c.Tor, c.TorControl)
		if err != nil {
			log.Printf("TOR: %v", err)
		}

		c.Collector.SetProxyFunc(pf)
	}
}

func (c *Crawler) crawl() {
	c.Collector.OnHTML(".contact_methods.hidden .contactitem strong", func(e *colly.HTMLElement) {
		fmt.Println(e.Text)
	})
	c.Collector.OnHTML(".lpv-item .lpv-item-link", func(e *colly.HTMLElement) {
		fmt.Println(e.Attr("href"))
		url := e.Request.AbsoluteURL(e.Attr("href"))
		c.Collector.Visit(url)
	})
	c.Collector.OnHTML(".site-map .link", func(e *colly.HTMLElement) {
		fmt.Println(e.Attr("href"))
		url := e.Request.AbsoluteURL(e.Attr("href"))
		c.Collector.Visit(url)
	})
}

// Start starts the crawler
func (c *Crawler) Start() {
	c.init()
	c.setProxy()
	c.Collector.SetRequestTimeout(c.Timeout)

	c.Collector.Limit(&colly.LimitRule{
		Parallelism: c.Parallelism,
		RandomDelay: c.RandomDelay,
		Delay:       c.Delay,
	})

	c.crawl()
	c.Collector.Visit(c.URL)
}
