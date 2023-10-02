package create_website

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/datastore"
)

const (
	WEBSITE_KIND = "websites"
	PROJECT_ID   = "PROJECT_ID"
)

type website struct {
	ID        *datastore.Key `json:"id" datastore:"-"`
	Title     string         `json:"title"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

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

	var website website
	if err := json.NewDecoder(r.Body).Decode(&website); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	now := time.Now()
	website.CreatedAt = now
	website.UpdatedAt = now

	key := datastore.IncompleteKey(WEBSITE_KIND, nil)
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

func createClient(ctx context.Context) (*datastore.Client, error) {
	client, err := datastore.NewClient(ctx, os.Getenv(PROJECT_ID))
	if err != nil {
		log.Printf("ERROR creating client: %v", err)
		return nil, err
	}

	return client, nil
}
