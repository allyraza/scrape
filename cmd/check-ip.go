package main

import (
	"flag"
	"log"
	"os"

	"github.com/allyraza/souqr/proxy"
	"github.com/gocolly/colly"
)

func main() {
	var (
		tor        = flag.String("tor", "http://localhost:8118", "tor proxy address")
		torControl = flag.String("tor-contrl", "localhost:9051", "tor control address")
		endpoint   = flag.String("endpoint", "http://httpbin.org/ip", "endpoint to check ip")
	)

	c := colly.NewCollector()

	pf, err := proxy.TorProxySwitcher(*tor, *torControl)
	if err != nil {
		log.Fatal(err)
	}

	c.SetProxyFunc(pf)

	c.OnResponse(func(r *colly.Response) {
		os.Stdout.Write(r.Body)
	})

	c.Visit(*endpoint)
}
