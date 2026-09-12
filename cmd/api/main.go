package main

import (
	"log"
	"net/http"
	"os"

	"github.com/k11ngp1ng/url-shortener-go/internal/httpapi"
	"github.com/k11ngp1ng/url-shortener-go/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := httpapi.NewServer(store.NewMemoryURLStore())
	log.Printf("URL Shortener API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, server.Routes()))
}
