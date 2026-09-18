package main

import (
	"fmt"
	"goscraper/scraper"
	"log"
	"net/url"
	"os"
)

func main() {
	log.SetFlags(0) //tolgo il timestamp

	arguments := os.Args[1:]

	if len(arguments) > 1 {
		log.Fatal("too many arguments provided")
	} else if len(arguments) == 0 {
		log.Fatal("no website provided")
	}

	fmt.Println("starting crawl of:", arguments[0])

	parsedBaseURL, err := url.Parse(arguments[0])

	if err != nil {
		log.Fatal(err)
	}

	data := make([]scraper.PageData, 0)
	pagesOcc := make(map[string]struct{})

	scraper.CrawlWebsite(parsedBaseURL, arguments[0], pagesOcc, &data)
}
