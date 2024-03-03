package main

import (
	"fmt"
	"strings"
	"os"
	"log"
	"sync"
	"encoding/csv"

	"github.com/gocolly/colly"
	"github.com/fatih/color"
)

var (
	blue = color.New(color.FgBlue, color.Bold).SprintFunc()
	white = color.New(color.FgWhite, color.Bold)
	green = color.New(color.FgGreen)
	green_line = color.New(color.FgGreen, color.Underline).SprintFunc()
)

type ScrapedData struct {
	URL	   string
	Source string
}

func request(url string, source string, wg *sync.WaitGroup, ch chan ScrapedData) {
	defer wg.Done()
	defer close(ch)

	c := colly.NewCollector()

	visited := make(map[string]bool)

	c.OnHTML("a", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if strings.HasPrefix(link, "https://") || strings.HasPrefix(link, "http://") {
			if !visited[link] {
				visited[link] = true
				ch <- ScrapedData{URL: link, Source: source}
				e.Request.Visit(link)
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Web-ScrapingBOT")

		white.Println("Visiting: ", r.URL)
		green.Printf("\n%s TARGET: %s\n", blue("[+]"), green_line(url))
	})

	c.Visit(url)
}

func main() {
	var wg sync.WaitGroup

	ch := make(chan ScrapedData)

	urlArg := os.Args[1:]
	if len(urlArg) == 0 {
	 	fmt.Printf("USE: ./%s URL1 URL2...\n", os.Args[0])
	 	return
	}

	urls := make(map[string]string)
	for _, url := range urlArg {
		name := extractUrlName(url)
		urls[url] = name
	}

	file, err := os.Create("result.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for url, source := range urls {
		wg.Add(1)
		go request(url, source, &wg, ch)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	headers := []string{"SOURCE", "SCRAPED URL"}

	var saveWG sync.WaitGroup
	saveWG.Add(1)
	go func() {
		defer saveWG.Done()
		writer.Write(headers)
		for data := range ch {
			err := writer.Write([]string{data.Source, data.URL})
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	saveWG.Wait()
}

func extractUrlName(url string) string {
	name := strings.Replace(url, "https://", "", 1)
	name = strings.Replace(name, "http://", "", 1)

	// Remove all after /
	parts := strings.Split(name, "/")
	name = parts[0]

	if strings.Contains(name, ".") {
		name = strings.Split(name, ".")[0]
	}

	return name
}
