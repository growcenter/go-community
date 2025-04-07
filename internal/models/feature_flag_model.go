package models

import "time"

var TYPE_FEATURE_FLAG = "featureFlag"

// FeatureFlag represents a feature flag in the system
type FeatureFlag struct {
	ID          int
	Name        string
	Key         string
	Description string
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Rules       *Rules `gorm:"type:jsonb;serializer:json"`
}

// Rules defines conditions for feature flag evaluation
type Rules struct {
	Percentage   *int                   `json:"percentage,omitempty"`    // Percentage of users who see the feature
	CommunityIds []string               `json:"community_ids,omitempty"` // Specific user IDs
	Parameters   map[string]interface{} `json:"parameters,omitempty"`    // Additional parameters
}

func (f *FeatureFlagResponse) ToResponse() *FeatureFlagResponse {
	return &FeatureFlagResponse{
		Type:        TYPE_FEATURE_FLAG,
		Name:        f.Name,
		Key:         f.Key,
		Description: f.Description,
		Enabled:     f.Enabled,
		Rules:       f.Rules,
	}
}

type (
	ToggleFlagRequest struct {
		Enabled bool `json:"enabled" validate:"required"`
	}
	FeatureFlagRequest struct {
		Name        string `json:"name" validate:"required"`
		Key         string `json:"key" validate:"required"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
		Rules       *Rules `json:"rules,omitempty"`
	}
	FeatureFlagResponse struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		Key         string `json:"key"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
		Rules       *Rules `json:"rules,omitempty"`
	}
)
