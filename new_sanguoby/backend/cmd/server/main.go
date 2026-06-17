package main

import (
	"log"
	"net/http"
	"os"

	"new_sanguoby/backend/internal/server"
)

func main() {
	addr := envOrDefault("HOST", "127.0.0.1") + ":" + envOrDefault("PORT", "8642")
	log.Printf("new_sanguoby backend listening on %s", addr)
	if err := http.ListenAndServe(addr, server.New()); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
