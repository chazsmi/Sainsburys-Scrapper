package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/browser"
)

type Products struct {
	Results []Product `json:"results"`
	Total   float64   `json:"total"`
}

type Product struct {
	Title       string  `json:"title"`
	Size        string  `json:"size"`
	UnitPrice   float64 `json:"unit_price"`
	Description string  `json:"description"`
}

func main() {
	// Process flag
	url := flag.String("url", "", "Add the starting url to scrap")
	flag.Parse()
	if *url == "" {
		log.Fatal("Please pass in a url to from")
	}

	var prods []Product

	chProducts := make(chan Product)
	count := 0

	bowIns := MakeRequest(*url)

	bowIns.Find(".productLister li").Each(func(i int, s *goquery.Selection) {
		count++
		link, ok := s.Find("h3 a").Attr("href")
		if ok {
			// Process concurrently only if we have a link
			go ProcessPage(link, chProducts)
		}
	})

	// Wait for responses on the channel
	c := 0
	total := 0.0
	for c < count {
		p := <-chProducts
		// Add to the total
		total += p.UnitPrice
		prods = append(prods, p)
		c++
	}

	// We are done so close the channel
	close(chProducts)

	products := Products{Results: prods, Total: total}

	ToJson(products, os.Stdout)
}

// Convert the struct to json and output to the console formated
func ToJson(products Products, output io.Writer) {
	j, err := json.Marshal(products)
	if err != nil {
		log.Fatal(err.Error())
	}
	var out bytes.Buffer
	json.Indent(&out, j, "", "\t")
	out.WriteTo(output)
}

func ProcessPage(link string, chProducts chan Product) {

	var (
		size        string
		description string
	)

	newpage := MakeRequest(link)

	title := newpage.Find(".productTitleDescriptionContainer h1").Text()

	r := regexp.MustCompile(`£(?P<a>[0-9]{1,3}.[0-9]{1,2})`)

	unitPrice := r.FindStringSubmatch(newpage.Find(".pricePerUnit").Text())

	if len(unitPrice) < 1 {
		log.Fatal("Couldn't parse price for " + title)
	}

	unitPriceFloat, _ := strconv.ParseFloat(unitPrice[1], 2)

	// Retrive the size adn description by looping through a section of the page
	newpage.Find(".productDataItemHeader").Each(func(i int, s *goquery.Selection) {
		if s.Text() == "Size" {
			size = s.Next().Find("p").Text()
		}
		if s.Text() == "Description" {
			description = s.Next().Find("p").Text()
		}
	})

	// Add all data to the struct
	p := Product{
		Title:       title,
		Size:        size,
		UnitPrice:   unitPriceFloat,
		Description: description,
	}

	// Send it through the channel
	chProducts <- p
}

func MakeRequest(url string) *browser.Browser {

	// Open a new browser needs to be a browser so run js
	bow := surf.NewBrowser()

	err := bow.Open(url)
	if err != nil {
		panic(err)
	}

	return bow
}
