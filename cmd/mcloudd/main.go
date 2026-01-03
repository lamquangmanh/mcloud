package mcloudd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"database/sql"
	"mcloud/internal/cert"
	"mcloud/internal/cluster"
	"mcloud/internal/config"
	"mcloud/internal/database"
	"mcloud/internal/grpc"
	"mcloud/pkg/logger"
)

func startHTTPServer(ctx context.Context, cfg *config.Config, conn *sql.DB) {
	// Set up HTTP handlers for REST API
	mux := http.NewServeMux()

	// Register cluster-related HTTP routes (e.g., /cluster/status)
	cluster.InitModule(mux, conn)

	// Start HTTP server for REST API
	addr := fmt.Sprintf("%s:%d", cfg.Manager.HttpHost, cfg.Manager.HttpPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("Starting HTTP server on %s", addr)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server ListenAndServe: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server Shutdown: %v", err)
	}
}

func startGRPCServer(ctx context.Context, cfg *config.Config, conn *sql.DB) {
	// Create directory for certificates if it doesn't exist
	// os.MkdirAll("internal/cert", 0700)

	// Generate or load CA certificate and key
	caCert, caKey, err := cert.GenerateCAV2(cfg.Security.CACertPath, cfg.Security.CAKeyPath)
	if err != nil {
		logger.Error("Generate CA error: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Manager.GrpcHost, cfg.Manager.GrpcPort)
	// Generate or load server certificate signed by CA
	err = cert.GenerateServerCert(
		caCert,
		caKey,
		addr,
		cfg.Security.ServerCertPath,
		cfg.Security.ServerKeyPath,
	)
	if err != nil {
		logger.Error("Generate server certificate error: %v", err)
	}

	// Start gRPC server with mutual TLS authentication
	logger.Info("Starting gRPC server on %s", addr)
	go func() {
		if err := grpc.StartGRPCServer(
			addr,
			cfg.Security.CACertPath,
			cfg.Security.ServerCertPath,
			cfg.Security.ServerKeyPath,
		); err != nil {
			logger.Error("gRPC server error: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down gRPC server...")
	// Note: Implement graceful shutdown for gRPC server if needed
}

// main is the entry point for the mcloudd server process.
// It loads configuration, initializes the database, sets up HTTP and gRPC servers, and starts serving requests.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Load configuration from file (YAML) and check for errors
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error("Failed to load config: %v", err)
	}
	logger.Info("Loaded config: %+v", cfg)

	// Initialize database connection and run migrations
	conn, err := database.Connect()
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
	}
	logger.Info("Database initialized and migrated: %+v", conn)

	// --- HTTP server setup ---
	go startHTTPServer(ctx, cfg, conn)

	// --- gRPC server setup ---
	go startGRPCServer(ctx, cfg, conn)

	// // Set up HTTP handlers for REST API
	// mux := http.NewServeMux()

	// // Register cluster-related HTTP routes (e.g., /cluster/init)
	// cluster.InitModule(mux, conn)

	// // Start HTTP server for REST API
	// addr := fmt.Sprintf("%s:%d", cfg.Manager.Host, cfg.Manager.Port)
	// server := &http.Server{
	// 	Addr:         addr,
	// 	Handler:      mux,
	// 	ReadTimeout:  5 * time.Second,
	// 	WriteTimeout: 10 * time.Second,
	// }

	// log.Fatal(server.ListenAndServe())
	// log.Println("mcloudd listening on", addr)

	// // --- gRPC server setup (runs after HTTP server exits) ---
	// // Create directory for certificates if it doesn't exist
	// os.MkdirAll("internal/cert", 0700)

	// // Generate or load CA certificate and key
	// caCert, caKey, err := cert.GenerateCAV2("internal/cert/ca.crt", "internal/cert/ca.key")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Generate or load server certificate signed by CA
	// err = cert.GenerateServerCert(
	// 	caCert,
	// 	caKey,
	// 	"127.0.0.1",
	// 	"internal/cert/server.crt",
	// 	"internal/cert/server.key",
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Start gRPC server with mutual TLS authentication
	// log.Fatal(
	// 	grpc.StartGRPCServer(
	// 		":9443",
	// 		"internal/cert/ca.crt",
	// 		"internal/cert/server.crt",
	// 		"internal/cert/server.key",
	// 	),
	// )
	
	<-ctx.Done()
	logger.Info("Shutting down gracefully, press Ctrl+C again to force")
}
