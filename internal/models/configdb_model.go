package models

import (
	"database/sql"
	"time"
)

var TYPE_CONFIG = "config"

type Config struct {
	ID         int
	Identifier string
	Key        string
	Value      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  sql.NullTime
}

func (cr *ConfigResponse) ToResponse() *ConfigResponse {
	return &ConfigResponse{
		Type:       TYPE_CONFIG,
		Identifier: cr.Identifier,
		Key:        cr.Key,
		Value:      cr.Value,
	}
}

type (
	ConfigResponse struct {
		Type       string `json:"type"`
		Identifier string `json:"identifier"`
		Key        string `json:"key"`
		Value      string `json:"value"`
	}
)

type (
	GetLocationsByCampusCodeResponse struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
)
