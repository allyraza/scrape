package main

import (
	"flag"
	"fmt"
	"time"

	souqr "github.com/allyraza/souqr"
)

var (
	defaultUserAgent = `Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36`
)

func main() {
	var (
		url         = flag.String("url", "http://uae.souq.com/ae-en/", "url to scrape")
		tor         = flag.String("tor", "proxy1.ams.local", "tor proxy url")
		timeout     = flag.Int("timeout", 60, "request timeout")
		parallelism = flag.Int("parallelism", 2, "parallelism for crawler")
		delay       = flag.Int("delay", 30, "crawler request delay")
		randomDelay = flag.Int("random-delay", 30, "crawler request delay")
		cacheDir    = flag.String("cache", "cache", "cache directory to save cache")
		allowed     = flag.String("allowed", "uae.souq.com", "allowed domains")
		filters     = flag.String("filters", "uae\\.souq\\.com/(|ae-en.+)$", "url filters")
		userAgent   = flag.String("user-agent", defaultUserAgent, "user agent")
		debug       = flag.Bool("debug", false, "enable debugging")
		redisURL    = flag.String("redis-url", ":6379", "redis server url")
	)
	flag.Parse()

	fmt.Println("Starting crawler ...")

	c := &souqr.Crawler{
		URL:         *url,
		Tor:         *tor,
		Timeout:     time.Duration(*timeout) * time.Second,
		Parallelism: *parallelism,
		Delay:       time.Duration(*delay) * time.Second,
		RandomDelay: time.Duration(*randomDelay) * time.Second,
		CacheDir:    *cacheDir,
		Allowed:     *allowed,
		Filters:     *filters,
		UserAgent:   *userAgent,
		Debug:       *debug,
		RedisURL:    *redisURL,
	}

	c.Start()
}
