package scraper

import (
	"encoding/json"
	"log"
	"os"
	"slices"
)

func WriteJSONReport(pages map[string]PageData) error {
	log.SetFlags(0)

	jsonFile, err := os.Create("report.json")

	defer func() {
		err = jsonFile.Close()

		if err != nil {
			log.Fatal(err)
		}
	}()

	if err != nil {
		return err
	}

	encoder := json.NewEncoder(jsonFile)

	encoder.SetIndent("", "  ")

	keys := make([]string, len(pages))

	var i int

	for k, _ := range pages {
		keys[i] = k
		i++
	}

	slices.Sort(keys)

	pgs := make([]PageData, len(pages))

	for j, k := range keys {
		pgs[j] = pages[k]
	}

	err = encoder.Encode(pgs)

	if err != nil {
		return err
	}

	return nil
}
