package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func scrapeWebsite(c *gin.Context) {
	// Implemente a lógica para fazer o scraping de um website
}

func main() {
	r := gin.Default()

	r.PUT("/websites/:id/scrape", scrapeWebsite)

	http.Handle("/", r)
}
