package get_website_data

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/datastore"
)

type Scrapes struct {
	ID              *datastore.Key  `json:"-" datastore:"-"`
	URL             string          `json:"url"`
	Name            string          `json:"name"`
	RowSelector     string          `json:"rowSelector"`
	ColumnSelectors []string        `json:"columnSelectors"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	Iterations      []WebsiteScrape `json:"iterations"`
}

type WebsiteScrape struct {
	ScrapeTime time.Time `json:"scrapeTime"`
	Data       []byte    `json:"data"`
}

type ScrapesDTO struct {
	ID              *datastore.Key     `json:"id" datastore:"-"`
	URL             string             `json:"url"`
	Name            string             `json:"name"`
	RowSelector     string             `json:"rowSelector"`
	ColumnSelectors []string           `json:"columnSelectors"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
	Iterations      []WebsiteScrapeDTO `json:"iterations"`
}

type WebsiteScrapeDTO struct {
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
		log.Printf("ERROR creating Datastore client, err: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	key, err := datastore.DecodeKey(scrapesID)
	if err != nil {
		log.Printf("ERROR decoding Scrapes ID: %v", err)
		http.Error(w, "Invalid Scrapes ID format. It should be a valid Datastore key.", http.StatusBadRequest)
		return
	}

	var scrapes Scrapes
	if err := client.Get(ctx, key, &scrapes); err != nil {
		log.Printf("ERROR fetching Scrapes: %s", err)
		http.Error(w, "Scrapes not found", http.StatusNotFound)
		return
	}
	log.Println("------------>", scrapes)
	scrapes.ID = key

	scrapesDTO := ScrapesDTO{
		ID:              scrapes.ID,
		URL:             scrapes.URL,
		Name:            scrapes.Name,
		RowSelector:     scrapes.RowSelector,
		ColumnSelectors: scrapes.ColumnSelectors,
		CreatedAt:       scrapes.CreatedAt,
		UpdatedAt:       scrapes.UpdatedAt,
		Iterations:      TransformWebsiteScrapesToDTO(scrapes.Iterations),
	}

	// Retorna o Scrapes obtido como resposta
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scrapesDTO)
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

// TransformWebsiteScrapesToDTO transforms a list of WebsiteScrape to a list of WebsiteScrapeDTO.
func TransformWebsiteScrapesToDTO(scrapes []WebsiteScrape) []WebsiteScrapeDTO {
	var dtoList []WebsiteScrapeDTO

	for _, scrape := range scrapes {
		dto := WebsiteScrapeDTO{
			ScrapeTime: scrape.ScrapeTime,
		}

		// Deserialize the Data field using the existing function
		dto.Data, _ = DeserializeBytesToScrape(scrape.Data)

		dtoList = append(dtoList, dto)
	}

	return dtoList
}

// DeserializeBytesToScrape deserializes a byte slice to a []map[string]string.
func DeserializeBytesToScrape(data []byte) ([]map[string]string, error) {
	var scrape []map[string]string
	err := json.Unmarshal(data, &scrape)
	if err != nil {
		return nil, err
	}
	return scrape, nil
}
