package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"

	"mcloud/internal/constant"
)

// GenerateServerCert generates a server certificate signed by the given CA and writes it to files.
//
// Parameters:
//   ca      - The CA certificate used to sign the server certificate
//   caKey   - The CA's private key
//   addr    - The server's IP address (used as the certificate's IP SAN)
//   certPath - File path to write the server certificate PEM
//   keyPath  - File path to write the server private key PEM
//
// Returns:
//   error - If any error occurs during key generation, certificate creation, or file writing
func GenerateServerCert(
	ca *x509.Certificate,
	caKey *rsa.PrivateKey,
	addr string,
	certPath string, 
	keyPath string,
) error {
	// Generate a new 4096-bit RSA private key for the server
	key, _ := rsa.GenerateKey(rand.Reader, 4096)
	// Generate a random serial number for the certificate
	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<62))

	// Create a certificate template for the server
	cert := &x509.Certificate{
		SerialNumber: serial, // unique serial number
		Subject: pkix.Name{
			CommonName: constant.AppServerName, // subject CN
		},
		NotBefore:   time.Now(), // valid from now
		NotAfter:    time.Now().Add(365 * 24 * time.Hour * 10), // valid for 10 years
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment, // allowed usages
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, // for server authentication
		IPAddresses: []net.IP{net.ParseIP(addr)}, // set IP SAN
	}

	// Create the certificate, signed by the CA
	der, err := x509.CreateCertificate(rand.Reader, cert, ca, &key.PublicKey, caKey)
	if err != nil {
		return err
	}

	// Write the certificate and private key to files in PEM format
	writePEM(certPath, "CERTIFICATE", der)
	writePEM(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))

	return nil
}
