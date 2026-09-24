package main

import (
	"bufio"
	"fmt"
	"goscraper/scraper"
	"log"
	"net/url"
	"os"
	"strconv"
	"sync"
)

func main() {
	log.SetFlags(0) //tolgo il timestamp

	arguments := os.Args[1:]

	if len(arguments) > 4 {
		log.Fatal("too many arguments provided")
	} else if len(arguments) == 0 {
		log.Fatal("no website provided")
	}

	fmt.Println("starting crawl of:", arguments[0])

	parsedBaseURL, err := url.Parse(arguments[0])

	if err != nil {
		log.Fatal(err)
	}

	maxGoRoutines := 5
	maxPages := 10

	if len(arguments) == 2 {
		maxGoRoutines, err = strconv.Atoi(arguments[1])

		if err != nil {
			log.Fatal(err)
		}
	}

	if len(arguments) == 3 {
		maxPages, err = strconv.Atoi(arguments[2])

		if err != nil {
			log.Fatal(err)
		}
	}

	cfg := scraper.Config{
		BaseUrl:            parsedBaseURL,
		Mu:                 new(sync.RWMutex),
		ConcurrencyControl: make(chan struct{}, maxGoRoutines),
		Wg:                 new(sync.WaitGroup),
		PagesData:          make(map[string]scraper.PageData),
		MaxPages:           maxPages,
		PagesOccs:          make(map[string]struct{}),
	}

	cfg.Wg.Add(1)
	scraper.CrawlWebsite(arguments[0], &cfg)
	cfg.Wg.Wait()

	err = scraper.WriteJSONReport(cfg.PagesData)

	if err != nil {
		log.Fatal(err)
	}

	if len(arguments) == 4 {

		if arguments[3] != "-s" {
			log.Fatal("unknown flag")
		}

		clear(cfg.PagesOccs)

		xmlFile, err := os.Create("sitemap.xml")
		bw := bufio.NewWriter(xmlFile)

		defer func() {
			err = bw.Flush()

			if err != nil {
				log.Fatal(err)
			}

			err = xmlFile.Close()

			if err != nil {
				log.Fatal(err)
			}
		}()

		scraper.Serialize(arguments[0], &cfg, bw, 0)
	}
}
