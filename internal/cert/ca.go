package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

// writePEM writes a PEM-encoded block to a file at the given path.
// path: file path to write to
// typ:  PEM block type (e.g., "CERTIFICATE", "RSA PRIVATE KEY")
// bytes: DER-encoded bytes to encode as PEM
func writePEM(path, typ string, bytes []byte) {
	f, _ := os.Create(path) // create or truncate the file
	defer f.Close()
	pem.Encode(f, &pem.Block{Type: typ, Bytes: bytes}) // write PEM block
}

// GenerateCA generates a new RSA private key and a self-signed X.509 CA certificate.
// It writes the certificate and key to the given file paths in PEM format, and also returns them as strings.
// certPath: file path to write the certificate PEM
// keyPath:  file path to write the private key PEM
// Returns: certificate PEM string, private key PEM string, and error (if any)
func GenerateCA(certPath string, keyPath string) (certPEM string, keyPEM string, err error) {
	// Generate a new 2048-bit RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Create a certificate template for a self-signed CA
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), // serial number for the certificate
		Subject: pkix.Name{
			Organization: []string{"MCloud"},
			CommonName:   "MCloud Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour * 10), // valid for 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true, // mark as CA
	}

	// Self-sign the certificate using the private key
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Write the certificate and key to files in PEM format
	writePEM(certPath, "CERTIFICATE", certDER)
	writePEM(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privateKey))

	// Encode certificate to PEM string
	certPEMBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM string
	keyPEMBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return string(certPEMBlock), string(keyPEMBlock), nil
}

func GenerateCAV2(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<62))

	cert := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"MCloud"},
			CommonName:   "MCloud Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	writePEM(certPath, "CERTIFICATE", der)
	writePEM(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))

	return cert, key, nil
}
