package create_website

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
)

const (
	WEBSITE_KIND = "websites"
	PROJECT_ID   = "web-scraping-hub"
)

type website struct {
	ID        *datastore.Key `json:"id" datastore:"-"`
	Title     string         `json:"title"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

func CreateWebsite(w http.ResponseWriter, r *http.Request) {
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
	client, err := datastore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("ERROR creating client: %v", err)
		return nil, err
	}

	return client, nil
}

func main() {
	http.HandleFunc("/websites", CreateWebsite)
	port := "8080"
	log.Printf("Listening on port %s...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
