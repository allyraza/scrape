package main

import (
    "fmt"
    "strings"
    "time"
    "flag"
    "github.com/PuerkitoBio/goquery"
    "github.com/gocolly/colly"
    // "github.com/gocolly/colly/proxy"
)

const (
    BaseUrl = "https://www.zoom.com.br/"
)

var (
    Debug = flag.Bool("debug", false, "enable debugging information.")
    Timeout = flag.Int("timeout", 1000, "request timeout.")
    Delay = flag.Int("delay", 60, "request delay.")
)

func main() {
    flag.Parse()

    var merchantName string
    leads := make([]Lead, 0)
    timeout := time.Duration(*Timeout) * time.Second
    delay := time.Duration(*Delay) * time.Second

    c := colly.NewCollector()
    m := colly.NewCollector()
    d := colly.NewCollector()
    l := colly.NewCollector()

    m.Limit(&colly.LimitRule{
        Parallelism: 100,
        Delay: delay,
    })

    d.Limit(&colly.LimitRule{
        Parallelism: 100,
        Delay: delay,
    })

    l.Limit(&colly.LimitRule{
        Parallelism: 100,
        Delay: delay,
    })

    c.Limit(&colly.LimitRule{
        Parallelism: 100,
        Delay: delay,
    })

    // c.CacheDir = "./cache"

    // proxy, err := proxy.RoundRobinProxySwitcher("socks5://92.222.74.221:80","socks5://110.77.188.220:54318")
    // if err != nil {
    //     log.Fatal(err)
    // }
    
    c.SetRequestTimeout(timeout)
    // c.SetProxyFunc(proxy)

    d.SetRequestTimeout(timeout)
    // d.SetProxyFunc(proxy)

    l.SetRequestTimeout(timeout)
    // l.SetProxyFunc(proxy)

    m.SetRequestTimeout(timeout)
    // m.SetProxyFunc(proxy)

    if (*Debug) {
        l.OnError(func(r *colly.Response, err error) {
            fmt.Println("List Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err, "\n\n")
        })

        c.OnError(func(r *colly.Response, err error) {
            fmt.Println("Main Request URL:", r.Request.URL, "failed with response:", r, "Status", r.StatusCode, "\nError:", err, "\n\n")
        })

        m.OnError(func(r *colly.Response, err error) {
            fmt.Println("Merchant Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err, "\n\n")
        })

        d.OnError(func(r *colly.Response, err error) {
            fmt.Println("Detail Request URL:", r.Request.URL, "failed with response:", string(r.StatusCode), "\nError:", err, "\n\n")
        })
    }

    c.OnHTML("a[class=cat-name]", func(e *colly.HTMLElement) {
        link := e.Request.AbsoluteURL(e.Attr("href"))
        if (*Debug) {
            fmt.Printf("Category: %s\n", link)
        }
        
        l.Visit(link)
    })

    l.OnHTML("a[class=name-link]", func(e *colly.HTMLElement) {
        link := e.Request.AbsoluteURL(e.Attr("href"))
        if (*Debug) {
            fmt.Printf("List: %s\n", link)
        }

        d.Visit(link)
    })

    d.OnHTML("ul.product-list li", func(e *colly.HTMLElement) {name := e.ChildAttr(".store img", "alt")
        link := e.Request.AbsoluteURL(e.ChildAttr("div.container", "data-lead") + "&logAndRedirect=1&redirReferer=")
        merchantName = name

        if (*Debug) {
            fmt.Printf("Name: %s\n", name)
            fmt.Printf("Link: %s\n", link)
        }

        m.Visit(link)
    })
  
    m.OnHTML("body", func(e *colly.HTMLElement) {
        lead := Lead {
            Name: merchantName,
            Url: e.Request.URL.String(),
        }

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

        if (!contains(leads, lead)) {
            leads = append(leads, lead)
            lead.Print()
        }
    })

    c.Visit(BaseUrl)
    m.Wait()
}