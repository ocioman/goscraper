package scraper

import (
	"net/url"
	"sync"
)

/*
Il channel e' bufferizzato, quando spawno una goroutine invio un valore nel channel, se il channel e' pieno
(numero max di goroutine raggiunto) allora <- rimane in attesa e la goroutine anche se e' stata spawnata non esegue.
Quando la goroutine ha terminato, riceve un valore dal channel e si libera uno slot nel buffer
*/

type Config struct {
	BaseUrl            *url.URL
	Mu                 *sync.RWMutex
	ConcurrencyControl chan struct{}
	Wg                 *sync.WaitGroup
	PagesData          map[string]PageData
	PagesOccs          map[string]struct{}
	MaxPages           int
}
