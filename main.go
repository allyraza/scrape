package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"io/ioutil"
	"log"

	"strings"
)

type Item struct {
	Name           string              `json:"name"`
	SubCategory    string              `json:"sub_category"`
	SubSubCategory string              `json:"sub_sub_category"`
	URL            string              `json:"url"`
	Filters        map[string][]string `json:"filters"`
}

func (i Item) String() string {
	bites, err := json.Marshal(i)
	if err != nil {
		log.Fatal(err)
	}
	return string(bites)
}

func main() {
	fmt.Println("beslist.nl")

	items := map[string]Item{}

	c := colly.NewCollector(
		colly.CacheDir("./cache"))

	sc := colly.NewCollector(
		colly.CacheDir("./cache"))

	ssc := colly.NewCollector(
		colly.CacheDir("./cache"))

	fc := colly.NewCollector(
		colly.CacheDir("./cache"))

	fc.OnHTML("details.columnsearch__accordion", func(el *colly.HTMLElement) {

		url := el.Request.URL.String()
		title := el.ChildText("summary > em")

		if title == "" {
			return
		}

		filters := []string{}
		el.ForEach("ul label", func(_ int, e *colly.HTMLElement) {
			filters = append(filters, e.Attr("title"))
		})

		item := items[url]
		item.Filters[title] = filters

		items[url] = item
	})

	// Category
	c.OnHTML(".categories-dropdown__menu .categories-dropdown__link", func(el *colly.HTMLElement) {
		name := strings.TrimSpace(el.Text)
		url := el.Attr("data-url")
		if url == "" {
			url = el.Attr("href")
		}
		url = el.Request.AbsoluteURL(url)

		item := Item{
			Name:    name,
			URL:     url,
			Filters: map[string][]string{},
		}

		items[url] = item

		fc.Visit(url)
		sc.Visit(url)

		fmt.Printf(".")
	})

	// Sub Category
	sc.OnHTML(".columnsearch__catlist", func(el *colly.HTMLElement) {
		name := el.ChildText(".catlist__label--active:not(.active__marker)")

		el.ForEach(".columnsearch__list .columnsearch__list .catlist__label:not(.catlist__label--active)", func(_ int, e *colly.HTMLElement) {
			subCategory := strings.TrimSpace(e.Text)
			url := e.Request.AbsoluteURL(e.Attr("href"))
			item := Item{
				Name:           name,
				SubCategory:    subCategory,
				SubSubCategory: "",
				URL:            url,
				Filters:        map[string][]string{},
			}
			items[url] = item

			fc.Visit(url)
			ssc.Visit(url)
		})
	})

	// Sub Sub Category
	ssc.OnHTML(".columnsearch__catlist", func(el *colly.HTMLElement) {
		name := el.ChildText(".catlist__label--active:not(.active__marker)")
		subCategory := el.ChildText(".catlist__label--active.active__marker")

		el.ForEach(".columnsearch__list .columnsearch__list .columnsearch__list .catlist__label:not(.catlist__label--active)", func(_ int, e *colly.HTMLElement) {
			subSubCategory := strings.TrimSpace(e.Text)
			url := e.Request.AbsoluteURL(e.Attr("href"))
			item := Item{
				Name:           name,
				SubCategory:    subCategory,
				SubSubCategory: subSubCategory,
				URL:            url,
				Filters:        map[string][]string{},
			}

			items[url] = item

			fc.Visit(url)
		})
	})

	c.Visit("http://beslist.nl")

	jsn, err := json.Marshal(items)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("data.json", jsn, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("done!\n")

}
