package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"mcloud/internal/cert"
	"mcloud/internal/cluster"
	"mcloud/internal/config"
	"mcloud/internal/database"
	"mcloud/internal/grpc"
)

// main is the entry point for the mcloudd server process.
// It loads configuration, initializes the database, sets up HTTP and gRPC servers, and starts serving requests.
func main() {
	// Load configuration from file (YAML) and check for errors
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loaded config:", cfg)

	// Initialize database connection and run migrations
	conn, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database initialized and migrated", conn)

	// Set up HTTP handlers for REST API
	mux := http.NewServeMux()

	// Register cluster-related HTTP routes (e.g., /cluster/init)
	cluster.InitModule(mux, conn)

	// Start HTTP server for REST API
	addr := fmt.Sprintf("%s:%d", cfg.Manager.Host, cfg.Manager.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
	log.Println("mcloudd listening on", addr)

	// --- gRPC server setup (runs after HTTP server exits) ---
	// Create directory for certificates if it doesn't exist
	os.MkdirAll("internal/cert", 0700)

	// Generate or load CA certificate and key
	caCert, caKey, err := cert.GenerateCAV2("internal/cert/ca.crt", "internal/cert/ca.key")
	if err != nil {
		log.Fatal(err)
	}

	// Generate or load server certificate signed by CA
	err = cert.GenerateServerCert(
		caCert,
		caKey,
		"127.0.0.1",
		"internal/cert/server.crt",
		"internal/cert/server.key",
	)
	if err != nil {
		log.Fatal(err)
	}

	// Start gRPC server with mutual TLS authentication
	log.Fatal(
		grpc.StartGRPCServer(
			":9443",
			"internal/cert/ca.crt",
			"internal/cert/server.crt",
			"internal/cert/server.key",
		),
	)
	// go func() {
	// 	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)

	// 	// Ensure TLS certificates are available
	// 	caCert, serverCert, serverKey, err := cert.EnsureCertificates(cfg.CertDir, grpcAddr)
	// 	if err != nil {
	// 		log.Fatal("Failed to ensure TLS certificates:", err)
	// 	}

	// 	// Start gRPC server
	// 	err = grpc.StartGRPCServer(grpcAddr, caCert, serverCert, serverKey)
	// 	if err != nil {
	// 		log.Fatal("Failed to start gRPC server:", err)
	// 	}
	// }()

	// select {} // Block forever
}
