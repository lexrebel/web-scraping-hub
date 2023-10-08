# Web Scraping Hub

Web Scraping Hub is a Golang monorepo project designed for educational purposes as a final requirement for a Distributed Software Architecture course. It serves as a web scraping hub deployed on Google Cloud Functions.

## Project Overview

Web Scraping Hub is a distributed system built in Go that allows you to perform web scraping tasks efficiently. It's designed to showcase distributed software architecture concepts, making it ideal for educational purposes.

## Features

- Web scraping from multiple sources.
- Scalable architecture using Google Cloud Functions.
- Simple and modular codebase.
- Easy deployment and testing.

## Getting Started

To get started with this project:

1. Clone the repository to your local machine.
```bash
git clone https://github.com/yourusername/web-scraping-hub.git
cd web-scraping-hub
```
2. Run the project locally.
```bash
go run main.go
```
This will start the project locally, and you can access it at http://localhost:8080

## Using the APIs

The project exposes APIs that you can interact with, both locally and when deployed on GCP. Below, you'll find examples of how to use the APIs for each environment:

### Local Development

When running the project locally, you can use the following endpoints with a mock ID. The ID for all requests is returned in the create-website function.

```bash
# Create a website (POST)
curl -X POST localhost:8080/create-website

# Update a website (PUT)
curl -X PUT localhost:8080/update-website/?id=a7bqv4TNzN4eL9ayp9su

# Scrape a website (PUT)
curl -X PUT localhost:8080/scrape-website/?id=a7bqv4TNzN4eL9ayp9su

# Get website data (GET)
curl -X GET localhost:8080/get-website-data/?id=a7bqv4TNzN4eL9ayp9su

# Export website data as CSV (GET)
curl -X GET localhost:8080/export-website-data/?id=a7bqv4TNzN4eL9ayp9su
```

### Deployed on GCP
When the project is deployed on GCP, the same five endpoints are available, but they are accessed through GCP Cloud Functions via a GCP API Gateway that requires an X-API-Key header.

```bash
# Create a website (POST)
curl -X POST https://web-scraping-hub-v4-7dd7ezdt.uc.gateway.dev/websites -H "X-API-Key: your_api_key_here

# Update a website (PUT)
curl -X PUT https://web-scraping-hub-v4-7dd7ezdt.uc.gateway.dev/websites/{website_id} -H "X-API-Key: your_api_key_here

# Scrape a website (PUT)
curl -X PUT https://web-scraping-hub-v4-7dd7ezdt.uc.gateway.dev/scrapes/{website_id} -H "X-API-Key: your_api_key_here

# Get website data (GET)
curl -X GET https://web-scraping-hub-v4-7dd7ezdt.uc.gateway.dev/data/{website_id} -H "X-API-Key: your_api_key_here

# Export website data as CSV (GET)
curl -X GET https://web-scraping-hub-v4-7dd7ezdt.uc.gateway.dev/export/{website_id} -H "X-API-Key: your_api_key_here"
```

Make sure to replace your_api_key_here with the actual API key required for authentication when deployed on GCP.
For more detailed information on the request and response formats, consult the OpenAPI Spec v2 file located in the ./api-gateway/gateway.yaml.

## License
This project is open-source and available under the MIT License. Feel free to use it for educational purposes or adapt it for your own projects.