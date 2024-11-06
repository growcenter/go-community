package generator

import (
	"fmt"
	"math/rand"
	"time"
)

func AccountNumber() (accountNumber string, err error) {
	// if campus.ID == 0 || coolCategory.ID == 0 {
	// 	return "", err
	// }

	// // Set current time with designated format
	// timeCreated := time.Now().Format("20060102150405")
	// // Combine both campus.ID and userId and count how many are there
	// countIds := len(fmt.Sprintf("%d%d", campus.ID, coolCategory.ID))

	// // We want the whole acc number digit to be 10, so it should reduce the digit of the time created
	// var finalTimeValue string
	// cutNumber := 10 - countIds
	// if cutNumber > 0 {
	// 	finalTimeValue = timeCreated[:cutNumber]
	// }

	// generated := fmt.Sprintf("%d%s%d", campus.ID, finalTimeValue, coolCategory.ID)
	// return generated, nil

	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	prefix := r.Intn(90-10+1) + 10
	number := r.Intn(1e10)
	check := (prefix*int(1e10) + number) % 97

	return fmt.Sprintf("%02d%010d%02d", prefix, number, check), nil
}

func generateBaseAccountNumber() string {
	randomNumber := rand.NewSource(time.Now().UnixNano())
	baseAccountNumber := fmt.Sprintf("%010d", randomNumber)
	return baseAccountNumber
}

// Function to calculate Luhn checksum
func CalculateLuhnChecksum(baseAccountNumber string) int {
	sum := 0
	double := false

	// Iterate through the digits from right to left
	for i := len(baseAccountNumber) - 1; i >= 0; i-- {
		digit := int(baseAccountNumber[i] - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	// Calculate the checksum that makes the total sum a multiple of 10
	checksum := (10 - (sum % 10)) % 10
	return checksum
}

// Function to generate a full account number with Luhn checksum
func LuhnAccountNumber() string {
	baseAccountNumber := generateBaseAccountNumber()
	checksum := CalculateLuhnChecksum(baseAccountNumber)
	return fmt.Sprintf("%s%d", baseAccountNumber, checksum)
}
