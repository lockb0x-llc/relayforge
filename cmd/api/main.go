package main

import (
	"log"
	"os"

	"github.com/lockb0x-llc/relayforge/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := api.NewServer()
	log.Printf("Starting RelayForge API server on port %s", port)
	
	if err := server.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}