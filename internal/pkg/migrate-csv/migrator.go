package migrate_csv

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	_ "github.com/lib/pq"
)

func migrateToCSV() {
	connStr := "host= port= dbname= user= password= sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	file, err := os.Open("cmd/api/user.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header row
	header, err := reader.Read()
	if err != nil {
		panic(err)
	}

	// Check if the header row contains the expected column names
	expectedHeader := []string{"id", "account_number", "name", "phone_number", "email", "password", "address", "state", "status", "role", "token", "created_at", "updated_at", "deleted_at"} // replace ... with the other column names
	if !reflect.DeepEqual(header, expectedHeader) {
		panic("unexpected CSV header")
	}

	// Read all records from the CSV file
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		// Extract the 14 values from the record
		id := record[0]
		accountNumber := record[1]
		name := record[2]
		phoneNumber := record[3]
		email := record[4]
		password := record[5]
		address := record[6]
		state := record[7]
		status := record[8]
		role := record[9]
		token := record[10]
		// Parse the timestamp values
		var createdAt, updatedAt, deletedAt sql.NullTime
		if record[11] != "" {
			t, err := time.Parse("2006-01-02 15:04:05", record[11])
			if err != nil {
				panic(err)
			}
			createdAt = sql.NullTime{Time: t, Valid: true}
		}
		if record[12] != "" {
			t, err := time.Parse("2006-01-02 15:04:05", record[12])
			if err != nil {
				panic(err)
			}
			updatedAt = sql.NullTime{Time: t, Valid: true}
		}
		if record[13] != "" {
			t, err := time.Parse("2006-01-02 15:04:05", record[13])
			if err != nil {
				panic(err)
			}
			deletedAt = sql.NullTime{Time: t, Valid: true}
		}

		// Insert the data into the database
		_, err = db.Exec("INSERT INTO event_users (id, account_number, name, phone_number, email, password, address, state, status, role, token, created_at, updated_at, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)", id, accountNumber, name, phoneNumber, email, password, address, state, status, role, token, createdAt, updatedAt, deletedAt)
		if err != nil {
			panic(err)
		} else {
			fmt.Println("CSV data migrated successfully")
		}
	}
}
