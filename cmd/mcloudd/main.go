package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"mcloud/internal/cluster"
	"mcloud/internal/config"
	"mcloud/internal/database"
)

func main() {
	// Load configuration and check for errors
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loaded config:", cfg)

	// Initialize database connection
	conn, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database initialized and migrated", conn)

	// Set up HTTP handlers
	mux := http.NewServeMux()

	// Register cluster-related routes
	cluster.InitModule(mux, conn)

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Manager.Host, cfg.Manager.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
	log.Println("mcloudd listening on", addr)
}
