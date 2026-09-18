package scraper

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func NormalizeUrl(url string) (string, error) {
	var normalized string

	if len(url) == 0 {
		return "", fmt.Errorf("error: url is empty")
	}

	normalized, _ = strings.CutPrefix(url, "http://")
	normalized, _ = strings.CutPrefix(normalized, "https://")

	if normalized[len(normalized)-1] == '/' {
		normalized = normalized[:len(normalized)-1]
	}

	return normalized, nil
}

func GetHeadingFromHTML(HTML string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(HTML))
	if err != nil {
		return "", err
	}

	if heading := doc.Find("h1").First(); heading.Length() > 0 {
		return strings.TrimSpace(heading.Text()), nil
	}

	if heading := doc.Find("h2").First(); heading.Length() > 0 {
		return strings.TrimSpace(heading.Text()), nil
	}

	return "", nil
}

func GetFirstParagraphFromHTML(HTML string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(HTML))

	if err != nil {
		return "", err
	}

	if par := doc.Find("p").First(); par.Length() > 0 {
		return strings.TrimSpace(par.Text()), nil
	}

	return "", nil
}

func GetUrlsFromHTML(HTML string, baseUrl *url.URL) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(HTML))

	if err != nil {
		return nil, err
	}

	urls := make([]string, 0)

	//tutti i tag anchor che contengono l'attributo HREF
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		extractedUrl, _ := s.Attr("href")

		var abs *url.URL

		abs, err = baseUrl.Parse(extractedUrl)

		urls = append(urls, abs.String())
	})

	if err != nil {
		return nil, err
	}

	if len(urls) < 0 {
		return nil, nil
	}

	return urls, nil
}

func GetImagesFromHTML(HTML string, baseUrl *url.URL) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(HTML))

	if err != nil {
		return nil, err
	}

	urls := make([]string, 0)

	//tutti i tag img che contengono l'attributo src
	doc.Find("img[src]").Each(func(_ int, s *goquery.Selection) {
		extractedUrl, _ := s.Attr("src")

		var abs *url.URL

		/*
			-se l'extracted URL e' relativo (solo path) lo metto come path di quello base
			-nel caso in cui l'url estratto dovesse essere assoluto e con un host diverso o uguale
			da quello base allora viene ritornato direttamente quello estratto
		*/
		abs, err = baseUrl.Parse(extractedUrl)

		urls = append(urls, abs.String())
	})

	if err != nil {
		return nil, err
	}

	if len(urls) < 0 {
		return nil, nil
	}

	return urls, nil
}

func ExtractPageData(HTML, pageURL string) (PageData, error) {
	var pd PageData
	var err error

	pd.PageUrl, err = url.Parse(pageURL)

	if err != nil {
		return PageData{}, err
	}

	pd.FirstParagraph, err = GetFirstParagraphFromHTML(HTML)

	if err != nil {
		return PageData{}, err
	}

	pd.Heading, err = GetHeadingFromHTML(HTML)

	if err != nil {
		return PageData{}, err
	}

	pd.OutGoingLinks, err = GetUrlsFromHTML(HTML, pd.PageUrl)

	if err != nil {
		return PageData{}, err
	}

	pd.ImageURLs, err = GetImagesFromHTML(HTML, pd.PageUrl)

	if err != nil {
		return PageData{}, err
	}

	return pd, nil
}

/*
Casi base: l'URL fornito ha un host diverso da quello di partenza || outgoing links e' vuoto || ho gia' visitato il link
Sviluppo del recursion tree: ogni outgoing link puo avere n figli che possono essere foglie o nodi (se si verifica uno dei casi base)
*/

/*
Page data lo passo come puntatore perche' se con l'append supero la capacity, la copia del puntatore punta a una nuova zona di memoria
e questa cosa non e' visibile al chiamante
*/

func CrawlWebsite(baseUrl *url.URL, rawCurrUrl string, pagesOcc map[string]struct{}, data *[]PageData) {
	parsedCurrUrl, err := url.Parse(rawCurrUrl)

	if err != nil {
		log.Fatal(err)
	}

	normalizedCurrUrl, err := NormalizeUrl(rawCurrUrl)

	if err != nil {
		log.Fatal(err)
	}

	if _, ok := pagesOcc[normalizedCurrUrl]; ok {
		return
	}

	if parsedCurrUrl.Host != baseUrl.Host {
		return
	}

	pagesOcc[normalizedCurrUrl] = struct{}{}

	rawHTML, err := GetHTML(rawCurrUrl)

	pgData, err := ExtractPageData(rawHTML, rawCurrUrl)

	if err != nil {
		log.Fatal(err)
	}

	*data = append(*data, pgData)

	for _, ol := range pgData.OutGoingLinks {
		CrawlWebsite(baseUrl, ol, pagesOcc, data)
	}
}
