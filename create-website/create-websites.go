// websites.go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"

)

func createWebsite(c *gin.Context) {
    // Implemente a lógica para criar um website
}

func updateWebsite(c *gin.Context) {
    // Implemente a lógica para atualizar um website
}

func scrapeWebsite(c *gin.Context) {
    // Implemente a lógica para fazer o scraping de um website
}

func main() {
    r := gin.Default()

    r.POST("/websites", createWebsite)
    r.PUT("/websites/:id", updateWebsite)
    r.PUT("/websites/:id/scrape", scrapeWebsite)

    http.Handle("/", r)
}
