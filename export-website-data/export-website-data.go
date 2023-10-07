package export_website_data

import (
	"context"
	"encoding/csv"
	"log"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	SCRAPES_COLLECTION = "scrapes"
)

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

// scrapeWebsiteExportCSV is the handler for exporting scraped data as CSV.
func ExportWebsiteData(w http.ResponseWriter, r *http.Request) {
	// Get the website ID to export CSV data from the request parameters
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

	// Create a new context for exporting CSV data
	ctx := context.Background()

	// Query the scraped data in the database based on the ID
	client, err := createClient(ctx)
	if err != nil {
		log.Printf("ERROR creating Firestore client, err: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	scrapesDocRef := client.Collection(SCRAPES_COLLECTION).Doc(websiteID)
	var scrapes Scrapes

	// Get the scrapes document
	snapshot, err := scrapesDocRef.Get(ctx)
	if err != nil {
		if status, ok := status.FromError(err); ok && status.Code() != codes.NotFound {
			log.Println("ERROR fetching document:", err)
			http.Error(w, "Website: "+websiteID+" not found", http.StatusNotFound)
			return
		}
	} else {
		err := snapshot.DataTo(&scrapes)
		if err != nil {
			log.Fatalf("Error decoding the document: %v", err)
		}
	}

	// Export scraped data as CSV
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=exported_data.csv")
	csvWriter := csv.NewWriter(w)

	// Write CSV header based on column selectors, with ScrapeTime as the first column
	header := make([]string, len(scrapes.ColumnSelectors)+1)
	header[0] = "ScrapeTime" // First column is ScrapeTime
	copy(header[1:], scrapes.ColumnSelectors)
	csvWriter.Write(header)

	// Write CSV data
	for _, iteration := range scrapes.Iterations {
		for _, data := range iteration.Data {
			row := make([]string, len(scrapes.ColumnSelectors)+1)
			row[0] = iteration.ScrapeTime.Format(time.RFC3339) // First column is ScrapeTime
			for i, colSelector := range scrapes.ColumnSelectors {
				row[i+1] = data[colSelector]
			}
			csvWriter.Write(row)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		log.Printf("Error writing CSV data: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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
