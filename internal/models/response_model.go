package models

type Health struct {
	Type	string		`json:"type"`
	Status	interface{}	`json:"status"`
}