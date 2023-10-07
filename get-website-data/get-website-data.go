package get_website_data

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/firestore"
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

// GetScrapesByID é a função HTTP para obter um Scrapes por ID.
func GetWebsiteData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	// Obtenha o ID Scrapes a ser obtido dos parâmetros da solicitação
	re := regexp.MustCompile(`/?id=(.+)`)
	match := re.FindStringSubmatch(r.RequestURI)
	scrapesID := ""
	if len(match) > 1 {
		scrapesID = match[1]
	}
	if scrapesID == "" {
		http.Error(w, "Scrapes ID is required", http.StatusBadRequest)
		return
	}

	// Consulta o Scrapes no banco de dados com base no ID
	client, err := createClient(ctx)
	if err != nil {
		log.Printf("ERROR creating Firestore client, err: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	scrapesDocRef := client.Collection(SCRAPES_COLLECTION).Doc(scrapesID)
	var scrapes Scrapes

	// Get the scrapes document
	snapshot, err := scrapesDocRef.Get(ctx)
	if err != nil {
		log.Println("ERROR fetching document:", err)
		http.Error(w, "Scrapes: "+scrapesID+" not found", http.StatusNotFound)
		return

	} else {
		err := snapshot.DataTo(&scrapes)
		if err != nil {
			log.Fatalf("Error decoding the document: %v", err)
		}
	}
	scrapes.ID = scrapesID

	// Return the extracted data as response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scrapes)
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
