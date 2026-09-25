package scraper

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func NormalizeUrl(rawUrl string) (string, error) {
	var normalized string

	if len(rawUrl) == 0 {
		return "", fmt.Errorf("error: url is empty")
	}

	parsed, err := url.Parse(rawUrl)

	if err != nil {
		return "", err
	}

	parsed.Fragment = ""
	parsed.RawFragment = ""

	rawUrl = parsed.String()

	normalized, _ = strings.CutPrefix(rawUrl, "http://")
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

	pd.pageUrl, err = url.Parse(pageURL)

	if err != nil {
		return PageData{}, err
	}

	pd.PageUrlString = pd.pageUrl.String()

	pd.FirstParagraph, err = GetFirstParagraphFromHTML(HTML)

	if err != nil {
		return PageData{}, err
	}

	pd.Heading, err = GetHeadingFromHTML(HTML)

	if err != nil {
		return PageData{}, err
	}

	pd.OutGoingLinks, err = GetUrlsFromHTML(HTML, pd.pageUrl)

	if err != nil {
		return PageData{}, err
	}

	pd.ImageURLs, err = GetImagesFromHTML(HTML, pd.pageUrl)

	if err != nil {
		return PageData{}, err
	}

	return pd, nil
}

/*
Casi base: l'URL fornito ha un host diverso da quello di partenza || outgoing links e' vuoto || ho gia' visitato il link
Sviluppo del recursion tree: ogni outgoing link puo avere n figli che possono essere foglie o nodi (se si verifica uno dei casi base)
*/

func CrawlWebsite(rawCurrUrl string, cfg *Config) {
	defer func() {
		<-cfg.ConcurrencyControl
		cfg.Wg.Done()
	}()

	cfg.ConcurrencyControl <- struct{}{}

	cfg.Mu.RLock()

	if len(cfg.PagesData) >= cfg.MaxPages {
		cfg.Mu.RUnlock()
		return
	}

	cfg.Mu.RUnlock()

	parsedCurrUrl, err := url.Parse(rawCurrUrl)

	if err != nil {
		log.Fatal(err)
	}

	normalizedCurrUrl, err := NormalizeUrl(rawCurrUrl)

	if err != nil {
		log.Fatal(err)
	}

	cfg.Mu.Lock()

	if _, ok := cfg.PagesOccs[normalizedCurrUrl]; ok {
		cfg.Mu.Unlock()
		return
	}

	cfg.PagesOccs[normalizedCurrUrl] = struct{}{}

	cfg.Mu.Unlock()

	if parsedCurrUrl.Host != cfg.BaseUrl.Host {
		return
	}

	rawHTML, err := GetHTML(rawCurrUrl)

	if err != nil {
		fmt.Println(err)
	}

	pgData, err := ExtractPageData(rawHTML, rawCurrUrl)

	if err != nil {
		log.Fatal(err)
	}

	cfg.Mu.Lock()

	cfg.PagesData[normalizedCurrUrl] = pgData

	cfg.Mu.Unlock()

	for _, ol := range pgData.OutGoingLinks {
		cfg.Wg.Add(1)
		go CrawlWebsite(ol, cfg)
	}
}

func Serialize(rawCurrUrl string, cfg *Config, ostream io.Writer, depth int) {
	if len(cfg.PagesOccs) >= cfg.MaxPages {
		return
	}

	normalizedUrl, err := NormalizeUrl(rawCurrUrl)

	if err != nil {
		log.Fatal(err)
	}

	if _, ok := cfg.PagesOccs[normalizedUrl]; ok {
		writeLeaf(rawCurrUrl, ostream, depth)
		return
	}

	cfg.PagesOccs[normalizedUrl] = struct{}{}

	pgData := cfg.PagesData[normalizedUrl]

	if len(pgData.OutGoingLinks) == 0 {
		writeLeaf(rawCurrUrl, ostream, depth)
		return
	}

	ws := strings.Repeat("\t", depth)

	node := fmt.Sprintf("%s<url loc=\"%s\">\n", ws, rawCurrUrl)

	_, err = io.WriteString(ostream, node)

	if err != nil {
		log.Fatal(err)
	}

	for _, ol := range pgData.OutGoingLinks {
		Serialize(ol, cfg, ostream, depth+1)
	}

	closing := fmt.Sprintf("%s</url>\n", ws)

	_, err = io.WriteString(ostream, closing)

	if err != nil {
		log.Fatal(err)
	}
}

func writeLeaf(rawCurrUrl string, ostream io.Writer, depth int) {
	ws := strings.Repeat("\t", depth)

	leaf := fmt.Sprintf("%s<url loc=\"%s\"/>\n", ws, rawCurrUrl)

	_, err := io.WriteString(ostream, leaf)

	if err != nil {
		log.Fatal(err)
	}
}
