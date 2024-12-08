package generator

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenerateHashCode(data string, length int) string {
	// Generate SHA-256 hash of the input data
	hash := sha256.New()
	hash.Write([]byte(data))
	hashBytes := hash.Sum(nil)
	// Convert hash to hex string and truncate to the specified length
	return hex.EncodeToString(hashBytes)[:length]
}
