package main

import (
	"fmt"
	"log"
	"net/http"

	create_website "github.com/lexrebel/web-scraping-hub/create-website"
	export_website_data "github.com/lexrebel/web-scraping-hub/export-website-data"
	get_website_data "github.com/lexrebel/web-scraping-hub/get-website-data"
	scrape_website "github.com/lexrebel/web-scraping-hub/scrape-website"
	update_website "github.com/lexrebel/web-scraping-hub/update-website"
)

func main() {
	http.HandleFunc("/create-website", create_website.CreateWebsite)
	http.HandleFunc("/update-website/", update_website.UpdateWebsite)
	http.HandleFunc("/scrape-website/", scrape_website.ScrapeWebsite)
	http.HandleFunc("/get-website-data/", get_website_data.GetWebsiteData)
	http.HandleFunc("/export-website-data/", export_website_data.ExportWebsiteData)

	port := "8080"
	log.Printf("Listening on port %s...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
