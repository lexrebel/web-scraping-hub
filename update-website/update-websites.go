package update_website

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

// WebsiteDTO represents the input data for updating a website.
type WebsiteDTO struct {
	ID              *datastore.Key `json:"id" datastore:"-"`
	URL             string         `json:"url"`             // Updated URL of the website
	Name            string         `json:"name"`            // Updated name or identifier for the website
	RowSelector     string         `json:"rowSelector"`     // Updated selector for rows on the webpage
	ColumnSelectors []string       `json:"columnSelectors"` // Updated list of selectors for columns within each row
}

// Website represents the structure for storing website data in Datastore.
type Website struct {
	*WebsiteDTO
	CreatedAt time.Time `json:"createdAt"` // Timestamp for when the website entry was created
	UpdatedAt time.Time `json:"updatedAt"` // Timestamp for when the website entry was last updated
}

// UpdateWebsite is the HTTP handler function for updating a website.
func UpdateWebsite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	// // Check if the provided ID is valid (string ID)
	// if websiteDTO.ID == "" {
	// 	http.Error(w, "Invalid website ID", http.StatusBadRequest)
	// 	return
	// }

	// Fetch the existing website entity based on ID
	// key := datastore.NameKey(COLLECTION, websiteDTO.ID, nil)
	var existingWebsite Website
	if err := client.Get(ctx, websiteDTO.ID, &existingWebsite); err != nil {
		log.Printf("ERROR fetching website: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Update the fields with new data
	existingWebsite.ID = websiteDTO.ID
	existingWebsite.URL = websiteDTO.URL
	existingWebsite.Name = websiteDTO.Name
	existingWebsite.RowSelector = websiteDTO.RowSelector
	existingWebsite.ColumnSelectors = websiteDTO.ColumnSelectors
	existingWebsite.UpdatedAt = time.Now()

	// Save the updated website entity back to Datastore
	if _, err := client.Put(ctx, websiteDTO.ID, &existingWebsite); err != nil {
		log.Printf("ERROR updating website: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingWebsite)
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
