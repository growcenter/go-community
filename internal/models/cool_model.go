package models

import "time"

var TYPE_COOL = "cool"

type Cool struct {
	ID         int
	Name       string
	CampusCode string
	Status     string
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}
