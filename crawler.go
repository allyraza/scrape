package souqr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/allyraza/souqr/proxy"
	"github.com/garyburd/redigo/redis"
	"github.com/gocolly/colly"
	//	"github.com/gocolly/redisstorage"
)

// Crawler wrapper around colly
type Crawler struct {
	URL         string
	Tor         string
	Pc          *colly.Collector
	Vc          *colly.Collector
	Timeout     time.Duration
	Parallelism int
	RandomDelay time.Duration
	Delay       time.Duration
	CacheDir    string
	Filters     string
	Allowed     string
	UserAgent   string
	Store       redis.Conn
	RedisURL    string
	Debug       bool
	AllowedOnly bool
	PerPage     int
	Stats       int
}

func (c *Crawler) init() {
	rc, err := redis.Dial("tcp", c.RedisURL)
	if err != nil {
		log.Fatalf("Could not connect to redis: %v", err)
	}

	c.Store = rc
	c.Stats = 0

	c.Pc = colly.NewCollector(
		// cache directory
		colly.CacheDir(c.CacheDir),
		// url filters
		colly.URLFilters(regexp.MustCompile(c.Filters)),
		// allowed domains
		colly.AllowedDomains(c.Allowed),
		// user agent
		colly.UserAgent(c.UserAgent),
	)

	if c.Debug {
		c.Pc.OnRequest(func(r *colly.Request) {
			log.Printf("REQUEST: %v", r.URL.String())
		})
		c.Pc.OnResponse(func(r *colly.Response) {
			log.Printf("RESPONSE: %v", r.StatusCode)
		})
	}
	//
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

	c.Pc.OnError(func(r *colly.Response, err error) {
		log.Printf(`
		URL: 		  %v
		RESPONSE: %#v
		Error:    %v`,
			r.Request.URL, r, err)
	})
}

func (c *Crawler) setProxy() {
	if c.Tor == "" {
		return
	}

	proxyFunc, err := proxy.TorProxySwitcher(c.Tor)
	if err != nil {
		log.Printf("PROXY: %v", err)
	}

	c.Pc.SetProxyFunc(proxyFunc)
}

func (c *Crawler) crawl() {
	if c.AllowedOnly {
		c.Pc.OnHTML(`script[type="application/ld+json"]`, func(e *colly.HTMLElement) {
			blob := strings.TrimSpace(e.Text)

			schema := &Category{}
			json.Unmarshal([]byte(blob), &schema)

			for _, v := range schema.Items {
				c.Pc.Visit(v.URL)
			}

			for p := 1; p <= schema.Total/c.PerPage; p++ {
				q := e.Request.URL.Query()
				q.Set("page", strconv.Itoa(p))
				q.Set("section", "2")
				e.Request.URL.RawQuery = q.Encode()

				c.Pc.Visit(e.Request.URL.String())
			}
		})

	} else {
		c.Pc.OnHTML("a[href]", func(e *colly.HTMLElement) {
			e.Request.Visit(e.Attr("href"))
		})
	}

	c.Vc.OnHTML(".product_content", c.handleHTML)
	c.Pc.OnHTML(".product_content", c.handleHTML)
	c.Pc.OnHTML(".size-stand .item-connection:not(.active)", c.visitVariant)
	c.Pc.OnHTML(".colors-block .has-tip:not(.active)", c.visitVariant)
}

func (c *Crawler) visitVariant(el *colly.HTMLElement) {
	url := el.Request.URL.String()
	parts := strings.Split((strings.Split(url, "/"))[4], "-")
	el.Response.Ctx.Put("id", parts[len(parts)-1])

	c.Vc.Request("GET", el.ChildAttr("a", "data-url"), nil, el.Response.Ctx, nil)
}

func (c *Crawler) handleHTML(el *colly.HTMLElement) {
	res := bytes.NewReader(el.Response.Body)
	doc, _ := goquery.NewDocumentFromReader(res)
	j1 := doc.Find(`script[type="application/ld+json"]`).Text()

	var j2 string
	doc.Find(`script[type="text/javascript"]`).Each(func(i int, s *goquery.Selection) {
		str := strings.TrimSpace(s.Text())

		if strings.HasPrefix(str, "var globalBucket") {
			j2 = str[18:]
		}
	})

	s := &Souq{}

	json.Unmarshal([]byte(j1), &s)
	json.Unmarshal([]byte(j2), &s)

	p, err := json.Marshal(s)
	if err != nil {
		log.Printf("JSON: %v", err)
	}

	c.Stats = c.Stats + 1

	c.Store.Do("LPUSH", "products", string(p))
}

func (c *Crawler) setTimeout() {
	c.Pc.SetRequestTimeout(c.Timeout)
}

func (c *Crawler) printStats() {
	fmt.Printf("\nTOTAL: %v\n", c.Stats)
}

// Start starts the crawler
func (c *Crawler) Start() {
	c.init()
	c.setProxy()
	c.setTimeout()

	c.Pc.Limit(&colly.LimitRule{
		Parallelism: c.Parallelism,
		RandomDelay: c.RandomDelay,
		Delay:       c.Delay,
	})

	c.Vc = c.Pc.Clone()

	c.crawl()
	c.Pc.Visit(c.URL)

	c.printStats()
}
