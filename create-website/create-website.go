package create_website

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
)

const (
	COLLECTION = "websites"
)

// WebsiteDTO represents the input data for creating a website.
type WebsiteDTO struct {
	URL             string   `json:"url"`             // URL of the website to scrape
	Name            string   `json:"name"`            // Name or identifier for the website
	RowSelector     string   `json:"rowSelector"`     // Selector for rows on the webpage
	ColumnSelectors []string `json:"columnSelectors"` // List of selectors for columns within each row
}

// Website represents the structure for storing website data in Datastore.
type Website struct {
	ID *datastore.Key `json:"id" datastore:"-"`
	*WebsiteDTO
	CreatedAt time.Time `json:"createdAt"` // Timestamp for when the website entry was created
	UpdatedAt time.Time `json:"updatedAt"` // Timestamp for when the website entry was last updated
}

// CreateWebsite is the HTTP handler function for creating a new website.
func CreateWebsite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		log.Printf("ERROR creating Datastore client, err: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var websiteDTO WebsiteDTO
	if err := json.NewDecoder(r.Body).Decode(&websiteDTO); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create an entity with the provided data
	website := Website{WebsiteDTO: &websiteDTO}

	// Set the timestamps for creation and update
	now := time.Now()
	website.CreatedAt = now
	website.UpdatedAt = now

	key := datastore.IncompleteKey(COLLECTION, nil)
	key, err = client.Put(ctx, key, &website)
	if err != nil {
		log.Printf("ERROR setting a website: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	website.ID = key
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(website)
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
