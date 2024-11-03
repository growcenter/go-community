package validator

import "go-community/internal/pkg/generator"

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
