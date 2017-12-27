package main

import (
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
        d.Visit(lead.Url)
    })

    d.OnHTML("body", func(e *colly.HTMLElement) {

        e.DOM.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
            href, _ := s.Attr("href")
            text := s.Text()
            var email string

            if (strings.Contains(href, "tel:"))  {
                phone := strings.Split(href, ":")[1]
                lead.Phone = strings.TrimSpace(phone)
            }

            if (strings.Contains(href, "@") || strings.Contains(text, "@")) {
                email = text
            }

            if (strings.Contains(href, "mailto:"))  {
                email = strings.Split(href, ":")[1]
            }

            if (len(email) > 0) {
                lead.Email = strings.TrimSpace(email)
            }
        })

        for _, l := range leads {
            if (l.Url == url) {
                lead.Print()
            }
        }
    })

    c.Visit(BaseUrl)
    c.Wait()
}