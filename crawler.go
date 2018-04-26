package souqr

import (
	"bytes"
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/allyraza/souqr/proxy"
	"github.com/garyburd/redigo/redis"
	"github.com/gocolly/colly"
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
}

// SouqAttributes represents product attrs in resp
type SouqAttributes struct {
	EAN           string `json:"Alternative_EANs,omitempty"`
	Brand         string `json:"Brand,omitempty"`
	Size          string `json:"Size,omitempty"`
	TargetedGroup string `json:"Targeted_Group,omitempty"`
}

// SouqProduct represents a product in resp
type SouqProduct struct {
	Name           string         `json:"name"`
	Brand          string         `json:"brand"`
	Category       string         `json:"category"`
	Currency       string         `json:"currencyCode"`
	Discount       float32        `json:"discount"`
	ID             string         `json:"id"`
	ItemID         int            `json:"id_item"`
	Price          float32        `json:"price"`
	Quantity       int            `json:"quantity"`
	Variant        string         `json:"variant,omitempty"`
	Attributes     SouqAttributes `json:"attributes,omitempty"`
	ParentCategory string         `json:"super_category"`
}

// SouqPageData represents a PageData map in resp
type SouqPageData struct {
	ItemID     int         `json:"ItemIDs,omitempty"`
	Category   string      `json:"channel_name,eomitempty"`
	Reviews    int         `json:"item_reviews,omitempty"`
	Name       string      `json:"item_title,omitempty"`
	PriceRange string      `json:"price_ranges,omitempty"`
	EAN        string      `json:"s_ean,omitempty"`
	SoldOut    string      `json:"sold_out,omitempty"`
	Rating     int         `json:"s_item_rating_total,omitempty"`
	AvgRating  string      `json:"s_item_rating_avg,omitempty"`
	Product    SouqProduct `json:"product"`
}

// Souq respresents data collected from resp
type Souq struct {
	Name        string       `json:"name"`
	Image       string       `json:"image"`
	Description string       `json:"description"`
	URL         string       `json:"url"`
	Language    string       `json:"s_language,omitempty"`
	Country     string       `json:"s_country,omitempty"`
	GTIN        string       `json:"gtin13,omitempty"`
	Color       string       `json:"color,omitempty"`
	Data        SouqPageData `json:"Page_Data"`
}

func (c *Crawler) init() {
	rc, err := redis.Dial("tcp", c.RedisURL)
	if err != nil {
		log.Fatalf("Could not connect to redis: %v", err)
	}

	c.Store = rc

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

	// storage := &redisstorage.Storage{
	// 	Address: c.RedisURL,
	// 	Prefix:  "souqr_",
	// }
	// defer storage.Client.Close()
	//
	// err = c.Pc.SetStorage(storage)
	// if err != nil {
	// 	panic(err)
	// }
	//
	// if err := storage.Clear(); err != nil {
	// 	log.Fatal(err)
	// }

	c.Pc.OnError(func(r *colly.Response, err error) {
		log.Printf(`
		URL: 		  %v
		RESPONSE: %v
		Error:    %v`,
			r.Request.URL, r, err)
	})
}

func (c *Crawler) setProxy() {
	proxyFunc, err := proxy.TorProxySwitcher(c.Tor)
	if err != nil {
		log.Printf("PROXY: %v", err)
	}

	c.Pc.SetProxyFunc(proxyFunc)
}

func (c *Crawler) crawl() {
	c.Pc.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.Pc.OnHTML(".pagination-next a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.Vc.OnHTML(".product_content", c.handleVariant)
	c.Pc.OnHTML(".product_content", c.handleProduct)
	c.Pc.OnHTML(".size-stand .item-connection:not(.active)", c.visitVariant)
	c.Pc.OnHTML(".colors-block .has-tip:not(.active)", c.visitVariant)
}

func (c *Crawler) visitVariant(el *colly.HTMLElement) {
	url := el.Request.URL.String()
	parts := strings.Split((strings.Split(url, "/"))[4], "-")
	el.Response.Ctx.Put("id", parts[len(parts)-1])

	c.Vc.Request("GET", el.ChildAttr("a", "data-url"), nil, el.Response.Ctx, nil)
}

func (c *Crawler) handleProduct(el *colly.HTMLElement) {
	res, err := c.parseResponse(el)
	if err != nil {
		return
	}
	c.Store.Do("LPUSH", "products", res)
}

func (c *Crawler) handleVariant(el *colly.HTMLElement) {
	res, err := c.parseResponse(el)
	if err != nil {
		return
	}
	id := el.Request.Ctx.Get("id")
	c.Store.Do("HSET", "variants", id, res)
}

func (c *Crawler) parseResponse(el *colly.HTMLElement) (string, error) {
	res := bytes.NewReader(el.Response.Body)
	doc, _ := goquery.NewDocumentFromReader(res)
	p1 := doc.Find(`script[type="application/ld+json"]`).Text()

	var p2 string
	doc.Find(`script[type="text/javascript"]`).Each(func(i int, s *goquery.Selection) {
		str := strings.TrimSpace(s.Text())

		if strings.HasPrefix(str, "var globalBucket") {
			p2 = str[18:]
		}
	})

	s := &Souq{}

	json.Unmarshal([]byte(p1), &s)
	json.Unmarshal([]byte(p2), &s)

	jsonb, err := json.Marshal(s)
	if err != nil {
		log.Printf("JSON: %v", err)
	}

	return string(jsonb), err
}

func (c *Crawler) setTimeout() {
	c.Pc.SetRequestTimeout(c.Timeout)
}

// Start starts the crawler
func (c *Crawler) Start() {
	c.init()
	// c.setProxy()
	c.setTimeout()

	c.Pc.Limit(&colly.LimitRule{
		Parallelism: c.Parallelism,
		RandomDelay: c.RandomDelay,
		Delay:       c.Delay,
	})

	c.Vc = c.Pc.Clone()

	c.crawl()
	c.Pc.Visit(c.URL)
}
