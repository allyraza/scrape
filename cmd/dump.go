package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/allyraza/souqr"
	"github.com/garyburd/redigo/redis"
)

// XMLWriter generates xml files
type XMLWriter struct {
	File *os.File
}

func (w *XMLWriter) Write(p []byte) error {
	s := &souqr.Souq{}

	err := json.Unmarshal(p, &s)
	if err != nil {
		fmt.Println(err)
	}

	var InStock int
	if s.PageData.Product.Quantity > 0 {
		InStock = 1
	} else {
		InStock = 0
	}

	categoryPath := fmt.Sprintf("%s > %s",
		s.PageData.Product.Category,
		s.PageData.Product.ParentCategory)

	prd := &souqr.Product{
		ID:           s.PageData.ItemID,
		Name:         s.Name,
		CategoryPath: categoryPath,
		Brand:        s.PageData.Product.Brand,
		Deeplink:     s.URL,
		Description:  s.Description,
		Image:        s.Image,
		Color:        s.Color,
		Gender:       s.PageData.Product.Attributes.TargetedGroup,
		Size:         s.PageData.Product.Attributes.Size,
		EAN:          s.PageData.EAN,
		Quantity:     s.PageData.Product.Quantity,
		InStock:      InStock,
		Price:        s.PageData.Product.Price,
		PriceFrom:    s.PageData.Product.Price,
	}

	prd.Write(w.File)

	return nil
}

func main() {
	var (
		filename = flag.String("filename", "", "name of the file to write data to")
		redisURL = flag.String("redis-url", ":6379", "redis server url")
		delay    = flag.Int("delay", 5, "timeout delay")
		poll     = flag.Bool("poll", false, "poll redis for products")
		debug    = flag.Bool("debug", false, "enable debugging")
	)
	flag.Parse()

	file, err := os.Create(*filename)
	if err != nil {
		log.Fatal(err)
	}

	xmlw := &XMLWriter{
		File: file,
	}

	fmt.Fprintf(xmlw.File, xml.Header+"<products>")

	conn, err := redis.Dial("tcp", *redisURL)
	if err != nil {
		log.Fatalf("Could not connect to redis server: %v", err)
	}

	for {
		res, err := redis.Bytes(conn.Do("LPOP", "products"))
		if err != nil && res != nil && *debug {
			log.Printf("%v", err)
		}

		if res == nil && *poll {
			time.Sleep(time.Duration(*delay) * time.Second)
			continue
		}

		if res == nil && !*poll {
			break
		}

		xmlw.Write(res)
	}

	fmt.Fprintf(xmlw.File, "</products>")
}
