package scrape_website

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/datastore"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

const (
	WEBSITES_COLLECTION = "websites"
	SCRAPES_COLLECTION  = "scrapes"
)

// Website represents a website in Datastore.
type Website struct {
	ID              *datastore.Key `json:"id" datastore:"-"`
	URL             string         `json:"url"`
	Name            string         `json:"name"`
	RowSelector     string         `json:"rowSelector"`
	ColumnSelectors []string       `json:"columnSelectors"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

type Scrapes struct {
	ID              *datastore.Key `json:"id" datastore:"-"`
	URL             string         `json:"url"`
	Name            string         `json:"name"`
	RowSelector     string         `json:"rowSelector"`
	ColumnSelectors []string       `json:"columnSelectors"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	Scrapes         []string       `json:"scrapes"`
}

type WebsiteScrape struct {
	ScrapeTime time.Time `json:"scrapeTime"`
	Data       []map[string]string
}

// ScrapeWebsite is the HTTP function for web scraping a website.
func ScrapeWebsite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	// Get the website ID to be web scraped from the request parameters
	re := regexp.MustCompile(`/?id=(.+)`)
	match := re.FindStringSubmatch(r.RequestURI)
	websiteID := ""
	if len(match) > 1 {
		websiteID = match[1]
	}
	if websiteID == "" {
		http.Error(w, "Website ID is required", http.StatusBadRequest)
		return
	}

	// Query the website in the database based on the ID
	client, err := createClient(ctx)
	if err != nil {
		log.Printf("ERROR creating Datastore client, err: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	key, err := datastore.DecodeKey(websiteID)
	if err != nil {
		log.Printf("ERROR decoding website ID: %v", err)
		http.Error(w, "Invalid Website ID format. It should be a valid Datastore key.", http.StatusBadRequest)
		return
	}

	var website Website
	if err := client.Get(ctx, key, &website); err != nil {
		log.Printf("ERROR fetching website: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	website.ID = key

	// Execute web scraping based on the website's data
	scrapedData, err := scrapeWebsite(ctx, client, website)
	if err != nil {
		http.Error(w, "Error scraping website: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the extracted data as response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scrapedData)
}

// createClient creates a new Datastore client.
func createClient(ctx context.Context) (*datastore.Client, error) {
	client, err := datastore.NewClient(ctx, "web-scraping-hub")
	if err != nil {
		log.Printf("ERROR creating client: %v", err)
		return nil, err
	}

	return client, nil
}

func scrapeWebsite(ctx context.Context, client *datastore.Client, website Website) ([]map[string]string, error) {
	log.Println("Website ID:", website.ID)
	log.Println("Website ID.ID:", website.ID.ID)

	// Context for chromedp
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// Slice to store the extracted data
	var scrapedData []map[string]string

	var rowNodes []*cdp.Node
	// Execute a page in headless browser
	err := chromedp.Run(ctx,
		chromedp.Navigate(website.URL),
		chromedp.WaitReady(website.ColumnSelectors[0]),
		chromedp.Nodes(website.RowSelector, &rowNodes, chromedp.ByQueryAll),
	)
	if err != nil {
		return nil, err
	}

	// Iterate through row elements and collect column data
	for _, rowNode := range rowNodes {
		// Initialize a map to store row data
		rowData := make(map[string]string)

		// Use CSS selectors to collect data from columns
		for _, colSelector := range website.ColumnSelectors {
			log.Println("Column searched:", colSelector)
			var columnData string

			// Execute a CSS selector to find column elements within the row
			err := chromedp.Run(ctx,
				chromedp.Text(colSelector, &columnData, chromedp.FromNode(rowNode)))
			if err != nil {
				return nil, err
			}
			rowData[colSelector] = columnData
			log.Println("Column obtained:", columnData)
		}

		// Add rowData to the slice of extracted data
		scrapedData = append(scrapedData, rowData)

		// Add a log to show the extracted data
		log.Printf("Data extracted from the row: %+v", rowData)
	}

	scrapeData := WebsiteScrape{
		ScrapeTime: time.Now(),
		Data:       scrapedData,
	}
	jsonScrapedData, err := serializeScrape(scrapeData)
	if err != nil {
		return nil, err

	}

	// Create a key using the WebsiteID as the ID for the Scrapes entity
	scrapesKey := datastore.IDKey(SCRAPES_COLLECTION, website.ID.ID, nil)

	// Try to fetch the existing Scrapes entity
	var scrapes Scrapes
	if err := client.Get(ctx, scrapesKey, &scrapes); err != nil {
		if err != datastore.ErrNoSuchEntity {
			log.Printf("ERROR fetching existing Scrapes: %v", err)
			return nil, err
		}
		// If the Scrapes entity doesn't exist, initialize it
		scrapes = Scrapes{
			ID:      website.ID,
			Scrapes: make([]string, 0), // Initialize the slice
		}
	}

	// Append the new scrape to the existing Scrapes
	scrapes.Scrapes = append(scrapes.Scrapes, jsonScrapedData)

	// Save the updated Scrapes to Datastore
	if _, err := client.Put(ctx, scrapesKey, &scrapes); err != nil {
		log.Printf("ERROR saving updated Scrapes: %v", err)
		return nil, err
	}

	// Return the extracted data
	return scrapedData, nil
}

// Serialize WebsiteScrape to JSON
func serializeScrape(scrape WebsiteScrape) (string, error) {
	data, err := json.Marshal(scrape)
	if err != nil {
		log.Printf("ERROR marshaling scrape: %v", err)
		return "", err
	}
	return string(data), nil
}
