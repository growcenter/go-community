package validator

import (
	"errors"
	"github.com/google/uuid"
	"go-community/internal/pkg/generator"
	"strings"
)

// Function to validate an account number with Luhn checksum
func LuhnAccountNumber(accountNumber string) bool {
	if len(accountNumber) < 10 {
		return false // Account number should be 10 digits (9 + 1 checksum)
	}

	baseAccountNumber := accountNumber[:len(accountNumber)-1]
	checksum := int(accountNumber[len(accountNumber)-1] - '0')

	calculatedChecksum := generator.CalculateLuhnChecksum(baseAccountNumber)
	return checksum == calculatedChecksum
}

func AccountNumberType(accountNumber string) (string, error) {
	input := strings.TrimSpace(accountNumber)

	_, err := uuid.Parse(input)
	if err == nil {
		return "guest", nil
	}

	if LuhnAccountNumber(input) {
		return "communityId", nil
	}

	return "", errors.New("invalid account number")
}
