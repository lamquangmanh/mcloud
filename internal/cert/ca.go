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

	"mcloud/internal/constant"
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

func ReadPEM(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
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

// GenerateCAV2 generates a new RSA private key and a self-signed X.509 CA certificate.
// Unlike GenerateCA, this version returns the parsed certificate and key objects directly,
// which is useful for immediately using them to sign other certificates.
// It uses a stronger 4096-bit RSA key and includes CRL signing capability.
//
// Parameters:
//   certPath - File path where the certificate PEM will be written
//   keyPath  - File path where the private key PEM will be written
//
// Returns:
//   - *x509.Certificate: Parsed CA certificate object (ready to sign other certs)
//   - *rsa.PrivateKey: Parsed RSA private key object
//   - error: Any error that occurred during generation or file writing
//
// Example Input:
//   certPath = "certs/ca.crt"
//   keyPath  = "certs/ca.key"
//
// Example Output (Success):
//   cert = &x509.Certificate{
//     SerialNumber: big.NewInt(4611686018427387904), // random 62-bit number
//     Subject: pkix.Name{
//       Organization: []string{"MCloud"},
//       CommonName:   "MCloud Root CA",
//     },
//     NotBefore: time.Now(),
//     NotAfter:  time.Now().Add(10 years),
//     KeyUsage:  x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
//     IsCA:      true,
//   }
//   key = &rsa.PrivateKey{...} // 4096-bit RSA key
//   err = nil
//
// Example Output (Error - Key Generation Failed):
//   cert = nil
//   key = nil
//   err = "crypto/rsa: message too long for RSA key size"
//
// Side Effect:
//   Creates two files on disk:
//   1. certPath: PEM-encoded certificate
//      -----BEGIN CERTIFICATE-----
//      MIIFazCCA1OgAwIBAgIIQB...
//      -----END CERTIFICATE-----
//
//   2. keyPath: PEM-encoded private key (4096-bit RSA)
//      -----BEGIN RSA PRIVATE KEY-----
//      MIIJKAIBAAKCAgEA3Z7f...
//      -----END RSA PRIVATE KEY-----
func GenerateCAV2(certPath string, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Generate a new 4096-bit RSA private key (stronger than the 2048-bit in GenerateCA)
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	// Generate a random serial number (62-bit integer) for the certificate
	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<62))

	// Create a certificate template for a self-signed CA
	cert := &x509.Certificate{
		SerialNumber: serial, // unique serial number
		Subject: pkix.Name{
			Organization: []string{constant.OrganizationName},
			CommonName:   constant.RootCACommonName,
		},
		NotBefore:             time.Now(),                             // valid from now
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // valid for 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign, // can sign certs and CRLs
		IsCA:                  true, // mark as a Certificate Authority
		BasicConstraintsValid: true, // basic constraints are valid
	}

	// Create the certificate in DER format, self-signed
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	// Write the certificate and private key to files in PEM format
	writePEM(certPath, "CERTIFICATE", der)
	writePEM(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))

	// Return the certificate template and key objects (not the DER bytes)
	// Note: The cert template is returned, not the parsed certificate
	return cert, key, nil
}
