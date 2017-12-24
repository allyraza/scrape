package main

import (
    "fmt"
    "strings"
    "errors"
    "log"
    "time"
    "flag"
    "net/url"
    "encoding/base64"
    "github.com/PuerkitoBio/goquery"
    "github.com/gocolly/colly"
    // "github.com/gocolly/colly/proxy"
)

const (
    BaseUrl = "https://www.zoom.com.br/"
    Timeout = time.Duration(500) * time.Second
)

var (
    Debug = flag.Bool("debug", false, "enable debugging information.")
)

func formatUrl(str string) (string, error) {
    if (strings.Contains(str, "javascript")) {
        return str, errors.New("invalid url")
    }

    u, err := url.Parse(str)
    if (err != nil) {
        log.Fatal(err)
    }

    i, _ := base64.StdEncoding.DecodeString(strings.Split(u.Path, "/")[2])
    u2, _ := url.Parse(string(i))
    return u2.Path, nil
}

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
    var merchantName string
    c := colly.NewCollector()
    m := colly.NewCollector()
    d := colly.NewCollector()
    l := colly.NewCollector()

    m.Limit(&colly.LimitRule{
        Parallelism: 100,
    })

    // c.CacheDir = "./cache"

    // proxy, err := proxy.RoundRobinProxySwitcher("socks5://92.222.74.221:80","socks5://110.77.188.220:54318")
    // if err != nil {
    //     log.Fatal(err)
    // }
    
    c.SetRequestTimeout(Timeout)
    // c.SetProxyFunc(proxy)

    d.SetRequestTimeout(Timeout)
    // d.SetProxyFunc(proxy)

    l.SetRequestTimeout(Timeout)
    // l.SetProxyFunc(proxy)

    m.SetRequestTimeout(Timeout)
    // m.SetProxyFunc(proxy)

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
        }

        e.DOM.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
            href, _ := s.Attr("href")
            text := s.Text()

            if (strings.Contains(href, "tel:"))  {
                phone := strings.Split(href, ":")[1]
                lead.Phone = strings.TrimSpace(phone)
            }

            mailLink := strings.Contains(href, "mailto:") || strings.Contains(href, "@" + e.Request.URL.Host)
            mailText := strings.Contains(href, "@" + e.Request.URL.Host) || strings.Contains(text, "@" + e.Request.URL.Host)
            if (mailLink || mailText)  {
                email := strings.Split(href, ":")[1]
                lead.Email = strings.TrimSpace(email)
            }

        })

        lead.Print()
    })

    c.Visit(BaseUrl)
    m.Wait()
}