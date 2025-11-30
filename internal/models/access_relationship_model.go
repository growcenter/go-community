package models

import (
	"database/sql"
	"time"
)

type AccessRelationship struct {
	Id          string       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ObjectType  string       `gorm:"column:object_type;type:varchar(50);not null" json:"object_type"`
	ObjectId    string       `gorm:"column:object_id;type:varchar(50);not null" json:"object_id"`
	Relation    string       `gorm:"column:relation;type:varchar(50);not null" json:"relation"`
	SubjectType string       `gorm:"column:subject_type;type:varchar(50);not null" json:"subject_type"`
	SubjectId   string       `gorm:"column:subject_id;type:varchar(50);not null" json:"subject_id"`
	CreatedAt   time.Time    `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt   sql.NullTime `gorm:"column:deleted_at" json:"deleted_at"`
}
