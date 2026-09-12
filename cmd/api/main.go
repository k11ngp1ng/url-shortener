package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/k11ngp1ng/url-shortener-go/internal/httpapi"
	"github.com/k11ngp1ng/url-shortener-go/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var urlStore store.URLStore = store.NewMemoryURLStore()
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		postgresStore, err := store.NewPostgresURLStore(context.Background(), databaseURL)
		if err != nil {
			log.Fatal(err)
		}
		defer postgresStore.Close()
		urlStore = postgresStore
	}

	limit := 60
	if value := os.Getenv("RATE_LIMIT_PER_MINUTE"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			limit = parsed
		}
	}
	server := httpapi.NewServer(urlStore).WithRateLimit(limit, time.Minute)
	log.Printf("URL Shortener API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, server.Routes()))
}
