package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func updateWebsite(c *gin.Context) {
	// Implemente a lógica para atualizar um website
}

func main() {
	r := gin.Default()

	r.PUT("/websites/:id", updateWebsite)

	http.Handle("/", r)
}
