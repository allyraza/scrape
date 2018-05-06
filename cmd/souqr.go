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
		proxy       = flag.String("proxy", "", "proxy url")
		tor         = flag.String("tor", "", "tor proxy url")
		torControl  = flag.String("tor-control", "", "tor control url")
		timeout     = flag.Int("timeout", 60, "request timeout period")
		parallelism = flag.Int("parallelism", 2, "parallelism for crawler")
		perPage     = flag.Int("per-page", 30, "items per page")
		delay       = flag.Int("delay", 30, "crawler request delay")
		randomDelay = flag.Int("random-delay", 60, "crawler request delay")
		cacheDir    = flag.String("cache", "cache", "cache directory to save cache")
		allowed     = flag.String("allowed", "uae.souq.com", "allowed domains")
		filters     = flag.String("filters", "uae\\.souq\\.com/(|ae-en.+)$", "url filters")
		userAgent   = flag.String("user-agent", defaultUserAgent, "user agent")
		debug       = flag.Bool("debug", false, "enable debugging")
		redisURL    = flag.String("redis-url", ":6379", "redis server url")
		allowedOnly = flag.Bool("allowed-only", false, "restrict to current url only")
	)
	flag.Parse()

	// Start the crawler with configuration
	fmt.Println("Starting crawler ...")

	c := &souqr.Crawler{
		URL:         *url,
		Proxy:       *proxy,
		Tor:         *tor,
		TorControl:  *torControl,
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
		AllowedOnly: *allowedOnly,
		PerPage:     *perPage,
	}

	c.Start()
}
