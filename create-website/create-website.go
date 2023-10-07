package create_website

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

const (
	COLLECTION = "websites"
)

type WebsiteDTO struct {
	URL             string   `json:"url" firestore:"url"`
	Name            string   `json:"name" firestore:"name"`
	RowSelector     string   `json:"rowSelector" firestore:"rowSelector"`
	ColumnSelectors []string `json:"columnSelectors" firestore:"columnSelectors"`
}

type Website struct {
	ID string `json:"id" firestore:"-"`
	*WebsiteDTO
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
}

func CreateWebsite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		log.Printf("ERROR creating Firestore client, err: %v", err)
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

	// Add the website to Firestore
	ref, _, err := client.Collection(COLLECTION).Add(ctx, website)
	if err != nil {
		log.Printf("ERROR adding a website to Firestore: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the ID in the website struct
	website.ID = ref.ID

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(website)
}

// createClient creates a new Datastore client.
func createClient(ctx context.Context) (*firestore.Client, error) {
	client, err := firestore.NewClient(ctx, "web-scraping-hub")
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}

	return client, nil
}
