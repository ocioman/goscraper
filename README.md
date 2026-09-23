# goscraper

## 🇮🇹 Italiano

`goscraper` è un piccolo crawler web scritto in Go che visita in modo concorrente le pagine di un sito a partire da un URL base e genera un report JSON con i dati estratti.

### Funzionalità

- Crawl ricorsivo dei link interni (stesso host del dominio iniziale)
- Limite al numero massimo di goroutine concorrenti
- Limite al numero massimo di pagine da visitare
- Estrazione per ogni pagina di:
  - URL
  - Primo heading (`h1`, fallback `h2`)
  - Primo paragrafo (`p`)
  - Link in uscita (`a[href]`)
  - URL immagini (`img[src]`)
- Output ordinato in `report.json`

### Requisiti

- Go `1.26`

### Avvio

```bash
go run . <website_url> [max_goroutines] [max_pages]
```

Esempio:

```bash
go run . https://example.com 8 50
```

Parametri:

- `website_url` (obbligatorio): URL iniziale da cui partire
- `max_goroutines` (opzionale, default `5`): limite di concorrenza
- `max_pages` (opzionale, default `10`): numero massimo di pagine da salvare nel report

### Output

Il programma genera `report.json` nella root del progetto con una lista di oggetti:

```json
{
  "url": "https://example.com/page",
  "heading": "Titolo",
  "first_paragraph": "Primo paragrafo...",
  "out_going_links": ["..."],
  "image_urls": ["..."]
}
```

### Dettagli tecnici implementativi

- **Gestione concorrenza**
  - `sync.WaitGroup` traccia il lavoro totale delle goroutine di crawl
  - Un canale bufferizzato `ConcurrencyControl chan struct{}` impone il numero massimo di crawl simultanei
  - `sync.RWMutex` protegge mappe condivise (`PagesData`, `PagesOccs`) e letture/scritture concorrenti

- **Strategia di crawl**
  - Funzione principale: `CrawlWebsite`
  - Scarta URL esterni rispetto all’host di partenza
  - Normalizza URL (rimozione schema `http/https` e slash finale) per deduplicazione
  - Evita visite duplicate con la mappa `PagesOccs`
  - Interrompe l’esplorazione al raggiungimento di `MaxPages`

- **Download e parsing HTML**
  - `GetHTML` esegue richieste HTTP GET con timeout di 10s e `User-Agent` custom
  - Valida status code (`< 400`) e `Content-Type` (`text/html`)
  - Parsing DOM con `goquery` per heading, paragrafo, link e immagini
  - Risoluzione dei link relativi tramite `baseUrl.Parse(...)`

- **Serializzazione report**
  - `WriteJSONReport` ordina le pagine per chiave URL normalizzata
  - Scrive un array JSON indentato per output stabile e leggibile

---

## 🇬🇧 English

`goscraper` is a small Go web crawler that concurrently visits pages of a website starting from a base URL and generates a JSON report with extracted data.

### Features

- Recursive crawl of internal links (same host as the starting domain)
- Maximum concurrent goroutine limit
- Maximum page visit limit
- Per-page extraction of:
  - URL
  - First heading (`h1`, fallback `h2`)
  - First paragraph (`p`)
  - Outgoing links (`a[href]`)
  - Image URLs (`img[src]`)
- Sorted output written to `report.json`

### Requirements

- Go `1.26`

### Run

```bash
go run . <website_url> [max_goroutines] [max_pages]
```

Example:

```bash
go run . https://example.com 8 50
```

Parameters:

- `website_url` (required): starting URL
- `max_goroutines` (optional, default `5`): concurrency limit
- `max_pages` (optional, default `10`): max number of pages stored in the report

### Output

The program creates `report.json` at the project root as a list of objects:

```json
{
  "url": "https://example.com/page",
  "heading": "Title",
  "first_paragraph": "First paragraph...",
  "out_going_links": ["..."],
  "image_urls": ["..."]
}
```

### Technical implementation details

- **Concurrency model**
  - `sync.WaitGroup` tracks total crawl goroutine work
  - A buffered channel `ConcurrencyControl chan struct{}` enforces max parallel crawls
  - `sync.RWMutex` protects shared maps (`PagesData`, `PagesOccs`) from concurrent access

- **Crawling strategy**
  - Main function: `CrawlWebsite`
  - Skips URLs outside the starting host
  - Normalizes URLs (removes `http/https` and trailing slash) for deduplication
  - Prevents duplicate visits via `PagesOccs`
  - Stops traversal once `MaxPages` is reached

- **HTML fetch and parsing**
  - `GetHTML` performs HTTP GET with 10s timeout and custom `User-Agent`
  - Validates status code (`< 400`) and `Content-Type` (`text/html`)
  - Uses `goquery` to parse headings, first paragraph, links, and images
  - Resolves relative URLs via `baseUrl.Parse(...)`

- **Report serialization**
  - `WriteJSONReport` sorts pages by normalized URL key
  - Writes pretty-printed JSON array for deterministic, readable output
