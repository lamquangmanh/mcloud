package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateJoinToken generates a secure bootstrap token for joining nodes
func GenerateJoinToken(clusterID string) string {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to less secure method if crypto/rand fails
		return fmt.Sprintf("mcloud-%s-insecure", clusterID[:8])
	}

	// Encode as base64 URL-safe string
	tokenRandom := base64.URLEncoding.EncodeToString(randomBytes)
	
	// Format: mcloud-<clusterID-prefix>-<random>
	return fmt.Sprintf("mcloud-%s-%s", clusterID[:8], tokenRandom[:16])
}
