package main

import (
	"flag"
	"fmt"

	"github.com/allyraza/scrape"
)

func main() {
	var (
		url      = flag.String("url", "", "url to scrape")
		debug    = flag.Bool("debug", false, "enable debugging")
		allowed  = flag.String("allowed", "", "allowed domain to scrape")
		cacheDir = flag.String("cache-dir", "./cache", "cache directory")
	)
	flag.Parse()

	fmt.Println("Starting crawler ...")

	c := &scrape.Crawler{
		URL:      *url,
		Debug:    *debug,
		Allowed:  *allowed,
		CacheDir: *cacheDir,
	}

	c.Start()
}
