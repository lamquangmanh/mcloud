package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// StartGRPCServer starts a secure gRPC server with mutual TLS authentication.
//
// Parameters:
//   addr       - The address to listen on (e.g., ":50051")
//   caCert     - Path to the CA certificate file (PEM format)
//   serverCert - Path to the server certificate file (PEM format)
//   serverKey  - Path to the server private key file (PEM format)
//
// Returns:
//   error - If any error occurs during setup or serving
func StartGRPCServer(addr string, caCert string, serverCert string, serverKey string) error {
	// Load the server's certificate and private key
	cert, _ := tls.LoadX509KeyPair(serverCert, serverKey)

	// Load the CA certificate to verify client certificates
	caBytes, _ := os.ReadFile(caCert)
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caBytes)

	// Configure TLS for the server
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},                     // server cert
		ClientAuth:   tls.RequireAndVerifyClientCert,             // require and verify client certs
		ClientCAs:    caPool,                                    // trusted CA pool
	}

	// Listen on the specified TCP address
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// Create a new gRPC server with TLS credentials
	grpcServer := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)

	fmt.Println("gRPC server listening on", addr)
	// Start serving incoming gRPC connections
	return grpcServer.Serve(lis)
}
