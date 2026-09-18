package scraper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func GetHTML(rawURL string) (string, error) {
	log.SetFlags(0)
	req, err := http.NewRequest("GET", rawURL, nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Scraper/1.0")

	cl := &http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := cl.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		log.SetFlags(0)
		err = res.Body.Close()

		if err != nil {
			log.Fatal(err)
		}
	}()

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("error: status code %d", res.StatusCode)
	} else if content := res.Header.Get("Content-Type"); !strings.Contains(content, "text/html") {
		return "", fmt.Errorf("error: unsupported content type %s", content)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		log.Fatal(err)
	}

	return string(body), nil
}
