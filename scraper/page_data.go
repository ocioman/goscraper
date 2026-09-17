package scraper

import "net/url"

type PageData struct {
	PageUrl        *url.URL
	Heading        string
	FirstParagraph string
	OutGoingLinks  []string
	ImageURLs      []string
}
