package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getWebsiteData(c *gin.Context) {
	// Implemente a lógica para obter dados de um website
}

func main() {
	r := gin.Default()

	r.GET("/data/:website_id", getWebsiteData)

	http.Handle("/", r)
}
