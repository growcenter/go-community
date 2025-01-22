package models

import (
	"database/sql"
	"time"
)

type UserRelation struct {
	ID                 int    `json:"id"`
	CommunityId        string `json:"communityId"`
	RelatedCommunityId string `json:"relatedCommunityId"`
	RelationshipType   string `json:"relationshipType"`
	CreatedAt          *time.Time
	UpdatedAt          *time.Time
	DeletedAt          sql.NullTime
}

type FamilyRelation struct {
	CommunityId string `json:"communityId"`
	Parents     []User `json:"parents"`
	Spouse      *User  `json:"spouse"`
	Children    []User `json:"children"`
}

type GetFamilyRelationDBOutput struct {
	CommunityId      string `json:"communityId"`
	Name             string `json:"name"`
	RelationshipType string `json:"relationshipType"`
}
