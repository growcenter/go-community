package generator

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func GenerateHashCode(data string, length int) string {
	// Generate SHA-256 hash of the input data
	hash := sha256.New()
	hash.Write([]byte(data))
	hashBytes := hash.Sum(nil)
	// Convert hash to hex string and truncate to the specified length
	return hex.EncodeToString(hashBytes)[:length]
}

// IdentifierCode generates a unique, human-readable identifier code
// Format: PREFIX-YYYYMM-XXXXX
// Example: EVT-202401-A3K9M
func IdentifierCode(ctx context.Context, encode string, dateTime time.Time, prefix string) (string, error) {
	base32Encoding := base32.NewEncoding(encode).WithPadding(base32.NoPadding)
	dateComponent := dateTime.Format("200601")

	// Random component (5 characters from Crockford's Base32)
	randomBytes := make([]byte, 4) // 4 bytes = ~6.4 Base32 characters, we'll take 5
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	randomComponent := base32Encoding.EncodeToString(randomBytes)[:5]

	// Combine components
	code := fmt.Sprintf("%s-%s-%s", strings.ToUpper(prefix), dateComponent, randomComponent)

	return code, nil
}

// Slug creates a URL-friendly slug from event title
// Format: event-title-yyyymm
// Example: homebase-gathering-202401
func Slug(title string, eventDate time.Time) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with hyphens
	slug = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == ' ', r == '_', r == '.', r == ',':
			return '-'
		default:
			return -1 // Remove character
		}
	}, slug)

	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Add date suffix for uniqueness
	dateSuffix := eventDate.Format("200601")

	// Limit slug length (max 50 chars before date)
	if len(slug) > 50 {
		slug = slug[:50]
		slug = strings.TrimRight(slug, "-")
	}

	return fmt.Sprintf("%s-%s", slug, dateSuffix)
}

// InstanceCode generates a unique instance code with event prefix
// Format: {EventCode}-{RandomSuffix}
// Example: EVT-202401-A3K9M-X7K9
// The random suffix provides 36^4 = 1,679,616 possible combinations per event
func InstanceCode(ctx context.Context, eventCode string, encodeKey string) (string, error) {
	// Use standard alphanumeric charset for readability
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const suffixLength = 4

	// Generate random suffix
	randomBytes := make([]byte, suffixLength)
	for i := range randomBytes {
		// Generate random index
		idx := rand.Intn(len(charset))
		randomBytes[i] = charset[idx]
	}

	// Combine event code with random suffix
	instanceCode := fmt.Sprintf("%s-%s", eventCode, string(randomBytes))

	return instanceCode, nil
}
