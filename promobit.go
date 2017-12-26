package main

import (
    "fmt"
    "log"
    "time"
    "flag"
    "strings"
    "github.com/gocolly/colly"
)

const (
    BaseUrl = "https://www.promobit.com.br/promocoes/lojas/"
)

var (
    Debug = flag.Bool("debug", false, "enable debugging information.")
    Timeout = flag.Int("timeout", 1000, "request timeout.")
    Delay = flag.Int("delay", 60, "request delay.")
)

type Lead struct {
    Name string
    Email string
    Phone string
    Url string
}

func (l Lead) Print() {
    fmt.Printf("%s,%s,%s,%s\n", l.Name, l.Url, l.Phone, l.Email)
}

func main() {
    flag.Parse()

    leads := make([]Lead, 0)
    timeout := time.Duration(*Timeout) * time.Second
    delay := time.Duration(*Delay) * time.Second

    c := colly.NewCollector()
    d := colly.NewCollector()

    c.Limit(&colly.LimitRule{
        Parallelism: 100,
        Delay: delay,
    })

    d.Limit(&colly.LimitRule{
        Parallelism: 100,
        Delay: delay,
    })
  
    c.SetRequestTimeout(timeout)
    d.SetRequestTimeout(timeout)

    if (*Debug) {
        c.OnError(func(r *colly.Response, err error) {
            log.Println("Request URL:", r.Request.URL, "failed with response:", r, "Status", r.StatusCode, "\nError:", err, "\n\n")
        })
        d.OnError(func(r *colly.Response, err error) {
            log.Println("Request URL:", r.Request.URL, "failed with response:", r, "Status", r.StatusCode, "\nError:", err, "\n\n")
        })
    }
    
    c.OnHTML(".stores-list .list-item", func(e *colly.HTMLElement) {
        link := e.Request.AbsoluteURL( e.ChildAttr("a", "href") )
        c.Visit(link)
    })

    c.OnHTML(".side-store", func(e *colly.HTMLElement) {
        lead := Lead{
            Name: "",
            Url: "",
        }

        lead.Name = strings.TrimSpace(e.ChildText("h3"))

        url := strings.TrimSpace(e.ChildText("h4"))
        if (len(url) > 0) {
            lead.Url = "http://" + url
        }

        leads = append(leads, lead)
        lead.Print()
    })

    c.Visit(BaseUrl)
    c.Wait()
}