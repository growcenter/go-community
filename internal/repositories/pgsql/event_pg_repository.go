package pgsql

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/logger"

	"gorm.io/gorm"
)

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	GetByCode(ctx context.Context, code string) (*models.Event, error)
	CheckEventByCodeOrSlug(ctx context.Context, code string, slug string) (bool, error)
	UpdatePartial(ctx context.Context, code string, updates *models.UpdateEventRequest) error
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

// Create creates a new event
// Returns error if the event could not be created
func (er *eventRepository) Create(ctx context.Context, event *models.Event) (err error) {
	logger.EnrichContext(ctx, "db_operation", "event.create")
	err = er.db.WithContext(ctx).Create(&event).Error
	if err != nil {
		logger.EnrichContextWith(ctx, map[string]any{
			"db_error":      true,
			"db_table":      "events",
			"error_message": err.Error(),
		})
	}
	return err
}

// GetByCode retrieves an event by its code
// Returns the event if found, or an error if not found or on database error
func (er *eventRepository) GetByCode(ctx context.Context, code string) (*models.Event, error) {
	logger.EnrichContextWith(ctx, map[string]any{
		"db_operation": "event.get_by_code",
		"lookup_code":  code,
	})

	var event models.Event
	err := er.db.WithContext(ctx).Where("code = ?", code).First(&event).Error
	if err != nil {
		logger.EnrichContextWith(ctx, map[string]any{
			"db_error":      true,
			"db_table":      "events",
			"event_found":   false,
			"error_message": err.Error(),
		})
		return nil, err
	}

	logger.EnrichContext(ctx, "event_found", true)
	return &event, nil
}

// CheckEventByCodeOrSlug checks if an event exists with the given code or slug
// Pass empty string for code or slug to skip checking that field
// Returns true if an event with the given code or slug exists
func (er *eventRepository) CheckEventByCodeOrSlug(ctx context.Context, code string, slug string) (bool, error) {
	query := er.db.WithContext(ctx).Model(&models.Event{})

	// Build dynamic query based on provided parameters
	if code != "" && slug != "" {
		query = query.Where("code = ? OR slug = ?", code, slug)
	} else if code != "" {
		query = query.Where("code = ?", code)
	} else if slug != "" {
		query = query.Where("slug = ?", slug)
	} else {
		// Neither code nor slug provided
		return false, nil
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// UpdatePartial performs a partial update on an event, only updating fields that are explicitly provided
// This is a pure data access method - it executes the update and returns success/failure
// The usecase layer is responsible for validation and fetching updated data if needed
func (er *eventRepository) UpdatePartial(ctx context.Context, code string, updates *models.UpdateEventRequest) error {
	logger.EnrichContextWith(ctx, map[string]any{
		"db_operation": "event.update_partial",
		"event_code":   code,
	})

	// Build update map from model
	updateMap, fieldCount := updates.ToUpdateMap()

	// If no fields to update, return early
	if fieldCount == 0 {
		logger.EnrichContext(ctx, "no_fields_to_update", true)
		return nil
	}

	logger.EnrichContext(ctx, "fields_to_update", fieldCount)

	// Execute update
	err := er.db.WithContext(ctx).
		Model(&models.Event{}).
		Where("code = ?", code).
		Updates(updateMap).
		Error

	if err != nil {
		logger.EnrichContextWith(ctx, map[string]any{
			"db_error":      true,
			"db_table":      "events",
			"error_message": err.Error(),
		})
		return err
	}

	logger.EnrichContextWith(ctx, map[string]any{
		"update_succeeded": true,
		"fields_updated":   fieldCount,
	})

	return nil
}
