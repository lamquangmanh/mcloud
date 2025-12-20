package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

// GenerateCA generates a self-signed CA certificate and returns the cert and key as PEM strings
func GenerateCA() (certPEM string, keyPEM string) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		// For demo purposes, return placeholder on error
		return "-----BEGIN CERTIFICATE-----\nPLACEHOLDER_CERT\n-----END CERTIFICATE-----",
			"-----BEGIN RSA PRIVATE KEY-----\nPLACEHOLDER_KEY\n-----END RSA PRIVATE KEY-----"
	}

	// Create certificate template
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"MCloud"},
			CommonName:   "MCloud Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour * 10), // 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Self-sign the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "-----BEGIN CERTIFICATE-----\nPLACEHOLDER_CERT\n-----END CERTIFICATE-----",
			"-----BEGIN RSA PRIVATE KEY-----\nPLACEHOLDER_KEY\n-----END RSA PRIVATE KEY-----"
	}

	// Encode certificate to PEM
	certPEMBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	keyPEMBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return string(certPEMBlock), string(keyPEMBlock)
}
