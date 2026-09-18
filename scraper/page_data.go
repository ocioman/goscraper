package scraper

import "net/url"

type PageData struct {
	pageUrl        *url.URL
	PageUrlString  string   `json:"url"`
	Heading        string   `json:"heading"`
	FirstParagraph string   `json:"first_paragraph"`
	OutGoingLinks  []string `json:"out_going_links"`
	ImageURLs      []string `json:"image_urls"`
}
