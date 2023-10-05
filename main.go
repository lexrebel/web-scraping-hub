package main

import (
	"fmt"
	"log"
	"net/http"

	create_website "github.com/lexrebel/web-scraping-hub/create-website"
	update_website "github.com/lexrebel/web-scraping-hub/update-website"
)

func main() {
	http.HandleFunc("/create-website", create_website.CreateWebsite)
	http.HandleFunc("/update-website/", update_website.UpdateWebsite)
	port := "8080"
	log.Printf("Listening on port %s...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
