package update_website

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

func UpdateWebsite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	re := regexp.MustCompile(`/?id=(.+)`)
	match := re.FindStringSubmatch(r.RequestURI)
	stringID := ""
	if len(match) > 1 {
		stringID = match[1]
	}
	if stringID == "" {
		http.Error(w, "Website ID is required", http.StatusBadRequest)
		return
	}
	log.Println("Updating website with ID:", stringID)

	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		log.Printf("ERROR creating Firestore client, err: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	websiteRef := client.Collection(COLLECTION).Doc(stringID)
	var websiteDTO WebsiteDTO
	if err := json.NewDecoder(r.Body).Decode(&websiteDTO); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedWebsite := Website{
		ID:         websiteRef.ID,
		WebsiteDTO: &websiteDTO,
		UpdatedAt:  time.Now(),
	}

	_, err = websiteRef.Set(ctx, updatedWebsite)
	if err != nil {
		log.Printf("ERROR updating website: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedWebsite)
}

func createClient(ctx context.Context) (*firestore.Client, error) {
	client, err := firestore.NewClient(ctx, "web-scraping-hub")
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}

	return client, nil
}
