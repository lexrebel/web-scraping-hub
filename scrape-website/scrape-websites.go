package scrape_website

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

const (
	WEBSITES_COLLECTION = "websites"
	SCRAPES_COLLECTION  = "scrapes"
)

// Website represents a website in Datastore.
type Website struct {
	ID              string    `json:"id" firestore:"-"`
	URL             string    `json:"url"`
	Name            string    `json:"name"`
	RowSelector     string    `json:"rowSelector"`
	ColumnSelectors []string  `json:"columnSelectors"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Scrapes struct {
	ID              string          `json:"id" firestore:"-"`
	URL             string          `json:"url"`
	Name            string          `json:"name"`
	RowSelector     string          `json:"rowSelector"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	ColumnSelectors []string        `json:"columnSelectors"`
	Iterations      []WebsiteScrape `json:"iterations"`
}

type WebsiteScrape struct {
	ScrapeTime time.Time           `json:"scrapeTime"`
	Data       []map[string]string `json:"data"`
}

// ScrapeWebsite is the HTTP function for web scraping a website.
func ScrapeWebsite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// Create a new context for the scraping operation
	ctx := context.Background()

	// Query the website in the database based on the ID
	client, err := createClient(ctx)
	if err != nil {
		log.Printf("ERROR creating Datastore client, err: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	websiteDocRef := client.Collection(WEBSITES_COLLECTION).Doc(websiteID)
	var website Website

	// Get the website document
	snapshot, err := websiteDocRef.Get(ctx)
	if err != nil {
		http.Error(w, "Error fetching website: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Check if the document exists
	if snapshot.Exists() {
		if err := snapshot.DataTo(&website); err != nil {
			http.Error(w, "Error decoding website data: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Website not found", http.StatusNotFound)
		return
	}
	website.ID = websiteDocRef.ID

	// Execute web scraping based on the website's data
	scrapedData, err := scrape(website)
	if err != nil {
		log.Printf("ERROR scraping website: %v", err)
		http.Error(w, "Error scraping website: "+err.Error(), http.StatusInternalServerError)
		return
	}

	scrapesDocRef := client.Collection(SCRAPES_COLLECTION).Doc(websiteID)
	var scrapes Scrapes

	// Get the scrapes document
	snapshot, err = scrapesDocRef.Get(ctx)
	if err != nil {
		if status, ok := status.FromError(err); ok && status.Code() != codes.NotFound {
			log.Println("ERROR fetching document:", err)
			http.Error(w, "Website: "+website.ID+" not found", http.StatusNotFound)
			return
		}
	} else {
		err := snapshot.DataTo(&scrapes)
		if err != nil {
			log.Fatalf("Erro ao decodificar o documento: %v", err)
		}
	}

	// Update the scrapes document
	scrapes.ID = website.ID
	scrapes.URL = website.URL
	scrapes.Name = website.Name
	scrapes.RowSelector = website.RowSelector
	scrapes.CreatedAt = website.CreatedAt
	scrapes.UpdatedAt = website.UpdatedAt
	scrapes.ColumnSelectors = website.ColumnSelectors
	scrapes.Iterations = append(
		scrapes.Iterations,
		WebsiteScrape{ScrapeTime: time.Now(), Data: scrapedData})
	scrapesDocRef.Set(ctx, scrapes)

	// Return the extracted data as response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scrapedData)
}

// createClient creates a new Datastore client.
func createClient(ctx context.Context) (*firestore.Client, error) {
	client, err := firestore.NewClient(ctx, "web-scraping-hub")
	if err != nil {
		log.Printf("ERROR creating client: %v", err)
		return nil, err
	}
	return client, nil
}

func scrape(
	website Website) ([]map[string]string, error) {
	log.Println("Website ID:", website.ID)

	// initializing a chrome instance
	chromeCtx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// Slice to store the extracted data
	var scrapedData []map[string]string
	log.Println("Before scrapedData:", scrapedData)

	var rowNodes []*cdp.Node
	// Execute a page in headless browser
	err := chromedp.Run(
		chromeCtx,
		chromedp.Navigate(website.URL),
		chromedp.WaitReady(website.ColumnSelectors[0]),
		chromedp.Nodes(website.RowSelector, &rowNodes, chromedp.ByQueryAll),
	)
	if err != nil {
		log.Println("Erro", err)
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
			err := chromedp.Run(chromeCtx,
				chromedp.Text(colSelector, &columnData, chromedp.FromNode(rowNode)))
			if err != nil {
				return nil, err
			}
			rowData[colSelector] = columnData
			log.Println("Column obtained:", columnData)
		}

		// Add rowData to the slice of extracted data
		scrapedData = append(scrapedData, rowData)
		log.Println("After scrapedData:", scrapedData)

		// Add a log to show the extracted data
		log.Printf("Data extracted from the row: %+v", rowData)
	}

	// Return the extracted data
	return scrapedData, nil
}
