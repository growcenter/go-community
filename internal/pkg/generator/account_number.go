package generator

import (
	"fmt"
	"go-community/internal/models"
	"time"
)

func AccountNumber(campus *models.Campus, id int) (string, error) {
	timeCreated := time.Now().Format("20060102150405")
	initialValue := fmt.Sprintf("%d%d", campus.ID, id)

	var finalValue string
	reminder := 10 - len(initialValue)
	if reminder > 0 {
		finalValue = timeCreated[:reminder]
	}

	accountNumber := fmt.Sprintf("%s%s", initialValue, finalValue)
	return accountNumber, nil
}
