package models

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type EventRegistrationRecord struct {
	ID                uuid.UUID
	Name              string
	Identifier        string
	CommunityId       string
	EventCode         string
	InstanceCode      string
	IdentifierOrigin  string
	CommunityIdOrigin string
	UpdatedBy         string
	Status            string
	RegisteredAt      time.Time
	VerifiedAt        time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         sql.NullTime
}
