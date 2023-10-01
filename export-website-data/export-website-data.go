package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func exportWebsiteData(c *gin.Context) {
	// Implemente a lógica para exportar dados de um website
}

func main() {
	r := gin.Default()

	r.GET("/data/:website_id/export", exportWebsiteData)

	http.Handle("/", r)
}
