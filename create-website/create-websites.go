package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
)

const WEBSITE_COLLECTION = "websites"

type website struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func main() {

	client, err := createClient()
	if err != nil {
		log.Printf("ERROR creating firestore client, err: %v", err)
	}

	r := gin.Default()
	r.POST("/websites", createWebsite(client))
	r.Run("localhost:9090")

}

func createClient() (*firestore.Client, error) {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "web-scraping-hub"}

	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Printf("ERROR initializing app: %v", err)
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Print("ERROR creating client")
		log.Fatal(err)
	}

	return client, nil
}

func createWebsite(client *firestore.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		var w website
		if err := c.BindJSON(&w); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		now := time.Now()
		w.CreatedAt = now
		w.UpdatedAt = now

		ref := client.Collection(WEBSITE_COLLECTION).NewDoc()
		_, err := ref.Set(c, map[string]interface{}{
			"title":     w.Title,
			"createdAt": w.CreatedAt,
			"updatedAt": w.UpdatedAt,
		})
		if err != nil {
			log.Printf("ERROR setting a website: %s", err)
			c.JSON(http.StatusInternalServerError, "")
			return
		}

		c.JSON(http.StatusCreated, w)
	}
}
