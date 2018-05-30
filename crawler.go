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
	tor "github.com/allyraza/souqr/proxy"
	"github.com/garyburd/redigo/redis"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
	//	"github.com/gocolly/redisstorage"
)

// Crawler wrapper around colly
type Crawler struct {
	URL         string
	Proxy       string
	Tor         string
	TorControl  string
	Cc          *colly.Collector
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

	c.Cc = colly.NewCollector(
		// cache directory
		colly.CacheDir(c.CacheDir),
		// url filters
		colly.URLFilters(regexp.MustCompile(c.Filters)),
		// allowed domains
		colly.AllowedDomains(c.Allowed),
		// user agent
		colly.UserAgent(c.UserAgent),
	)

	c.Pc = c.Cc.Clone()
	c.Vc = c.Cc.Clone()

	if c.Debug {
		c.Cc.OnRequest(func(r *colly.Request) {
			log.Printf("REQUEST: %v", r.URL.String())
		})
		c.Cc.OnResponse(func(r *colly.Response) {
			log.Printf("RESPONSE: %v", r.StatusCode)
		})

		c.Pc.OnRequest(func(r *colly.Request) {
			log.Printf("REQUEST: %v", r.URL.String())
		})
		c.Pc.OnResponse(func(r *colly.Response) {
			log.Printf("RESPONSE: %v", r.StatusCode)
		})

		c.Vc.OnRequest(func(r *colly.Request) {
			log.Printf("REQUEST: %v", r.URL.String())
		})
		c.Vc.OnResponse(func(r *colly.Response) {
			log.Printf("RESPONSE: %v", r.StatusCode)
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

	c.Cc.OnError(func(r *colly.Response, err error) {
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

		c.Cc.SetProxyFunc(pf)
		c.Pc.SetProxyFunc(pf)
		c.Vc.SetProxyFunc(pf)
	}

	if c.Proxy != "" {
		proxies := strings.Split(c.Proxy, ",")

		pf, err := proxy.RoundRobinProxySwitcher(proxies...)
		if err != nil {
			log.Printf("Proxy: %v", err)
		}

		c.Cc.SetProxyFunc(pf)
		c.Pc.SetProxyFunc(pf)
		c.Vc.SetProxyFunc(pf)
	}
}

func (c *Crawler) crawl() {
	c.Cc.OnHTML(".side-nav li > a[href]", func(e *colly.HTMLElement) {
		exists, err := redis.Bool(c.Store.Do("SISMEMBER", "categories", e.Attr("href")))
		if err != nil {
			log.Println(err)
		}

		if !exists {
			log.Println(e.Attr("href"))
			c.Cc.Visit(e.Attr("href"))
			c.Store.Do("SADD", "categories", e.Attr("href"))
		}
	})

	c.Cc.OnHTML(`script[type="application/ld+json"]`, c.categoryHandler)
	c.Vc.OnHTML(".product_content", c.detailHandler)
	c.Pc.OnHTML(".product_content", c.detailHandler)
	c.Pc.OnHTML(".size-stand .item-connection:not(.active)", c.variantVisitor)
	c.Pc.OnHTML(".colors-block .has-tip:not(.active)", c.variantVisitor)
}

func (c *Crawler) variantVisitor(el *colly.HTMLElement) {
	url := el.Request.URL.String()
	parts := strings.Split((strings.Split(url, "/"))[4], "-")
	el.Response.Ctx.Put("id", parts[len(parts)-1])

	c.Vc.Request("GET", el.ChildAttr("a", "data-url"), nil, el.Response.Ctx, nil)
}

func (c *Crawler) categoryHandler(e *colly.HTMLElement) {
	blob := strings.TrimSpace(e.Text)

	schema := &Category{}
	json.Unmarshal([]byte(blob), &schema)

	for _, v := range schema.Items {
		c.Pc.Visit(v.URL)
	}

	totalPages := schema.Total / c.PerPage

	for p := 1; p <= totalPages; p++ {
		q := e.Request.URL.Query()
		q.Set("page", strconv.Itoa(p))
		q.Set("section", "2")
		e.Request.URL.RawQuery = q.Encode()

		if q.Get("redirectup") == "1" {
			break
		}

		c.Cc.Visit(e.Request.URL.String())
	}
}

func (c *Crawler) detailHandler(el *colly.HTMLElement) {
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
	category := strings.Join(strings.Fields(s.PageData.Product.Category), "_")
	category = strings.ToLower(strings.Replace(category, "&", "and", -1))

	c.Store.Do("LPUSH", category, string(p))
	c.Store.Do("SET", "current", category)

	fmt.Printf("\rProducts: %v", c.Stats)
}

func (c *Crawler) setTimeout() {
	c.Cc.SetRequestTimeout(c.Timeout)
	c.Pc.SetRequestTimeout(c.Timeout)
	c.Vc.SetRequestTimeout(c.Timeout)
}

// Start starts the crawler
func (c *Crawler) Start() {
	c.init()
	c.setProxy()
	c.setTimeout()

	c.Cc.Limit(&colly.LimitRule{
		Parallelism: c.Parallelism,
		RandomDelay: c.RandomDelay,
		Delay:       c.Delay,
	})

	c.Pc.Limit(&colly.LimitRule{
		Parallelism: c.Parallelism,
		RandomDelay: c.RandomDelay,
		Delay:       c.Delay,
	})

	c.Vc.Limit(&colly.LimitRule{
		Parallelism: c.Parallelism,
		RandomDelay: c.RandomDelay,
		Delay:       c.Delay,
	})

	c.crawl()
	c.Cc.Visit(c.URL)
}
