package pgsql

import (
	"context"
	"errors"
	"go-community/internal/models"
	"go-community/internal/pkg/logger"

	"gorm.io/gorm"
)

type EventRepository interface {
	// Create
	Create(ctx context.Context, event *models.Event) error

	// Get
	GetDummy(ctx context.Context) ([]models.Event, error)
	GetByCode(ctx context.Context, code string) (*models.Event, error)

	// Check
	CheckByCode(ctx context.Context, code string) (bool, error)
	CheckBySlug(ctx context.Context, slug string) (bool, error)
	CheckByCodeOrSlug(ctx context.Context, identifier ...string) (bool, error)

	// Update
	UpdatePartial(ctx context.Context, code string, updates *models.UpdateEventRequest) error
}

type eventRepository struct {
	*BaseRepository
}

// Compile-time interface compliance check
var _ EventRepository = (*eventRepository)(nil)

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new event
// Returns error if the event could not be created
func (er *eventRepository) Create(ctx context.Context, event *models.Event) (err error) {
	logger.AddProcess(ctx, "db_operation", "event.create")
	err = er.db(ctx).Create(&event).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_CREATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "events",
			},
		})
	}
	return err
}

// GetDummy retrieves all event without filter
// Returns the event if found, or an error if not found or on database error
func (er *eventRepository) GetDummy(ctx context.Context) ([]models.Event, error) {
	logger.AddProcess(ctx, "db_operation", "event.get_dummy")

	var event []models.Event
	err := er.db(ctx).Find(&event).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "events",
			},
		})
		return nil, err
	}

	logger.Add(ctx, "event_found", true)
	return event, nil
}

// GetByCode retrieves an event by its code
// Returns the event if found, or an error if not found or on database error
func (er *eventRepository) GetByCode(ctx context.Context, code string) (*models.Event, error) {
	logger.AddProcess(ctx, "db_operation", "event.get_by_code")

	var event models.Event
	err := er.db(ctx).Where("code = ?", code).First(&event).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "events",
			},
		})
		return nil, err
	}

	logger.Add(ctx, "event_found", true)
	return &event, nil
}

// CheckByCode checks if a session code exists in the database
// Uses PostgreSQL EXISTS for optimal performance - stops at first match instead of counting all rows
// Returns true if the code exists, false otherwise
func (er *eventRepository) CheckByCode(ctx context.Context, code string) (bool, error) {
	logger.AddProcess(ctx, "db_operation", "event.check_by_code")

	var exists bool
	err := er.db(ctx).
		Raw(QueryCheckEventByCode, code).
		Scan(&exists).Error

	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_CHECK_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "events",
			},
		})
		return false, err
	}

	logger.Add(ctx, "code_exists", exists)
	return exists, nil
}

// CheckBySlug checks if a session slug exists in the database
// Uses PostgreSQL EXISTS for optimal performance - stops at first match instead of counting all rows
// Returns true if the slug exists, false otherwise
func (er *eventRepository) CheckBySlug(ctx context.Context, slug string) (bool, error) {
	logger.AddProcess(ctx, "db_operation", "event.check_by_slug")

	var exists bool
	err := er.db(ctx).
		Raw(QueryCheckEventBySlug, slug).
		Scan(&exists).Error

	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_CHECK_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "events",
			},
		})
		return false, err
	}

	logger.Add(ctx, "slug_exists", exists)
	return exists, nil
}

// CheckByCodeOrSlug checks if an event exists with the given code or slug
// Pass empty string for code or slug to skip checking that field
// Returns true if an event with the given code or slug exists
func (er *eventRepository) CheckByCodeOrSlug(ctx context.Context, identifier ...string) (bool, error) {
	logger.AddProcess(ctx, "db_operation", "event.check_by_code_or_slug")

	var exists bool
	var err error
	if len(identifier) == 2 {
		code := identifier[0]
		slug := identifier[1]
		err = er.db(ctx).
			Raw(QueryCheckEventByCodeOrSlug, code, slug).
			Scan(&exists).Error
	} else if len(identifier) == 1 {
		code := identifier[0]

		err = er.db(ctx).
			Raw(QueryCheckEventByCodeOrSlug, code, code).
			Scan(&exists).Error
	} else {
		return false, errors.New("should have 1 or 2 identifiers")
	}

	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_CHECK_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "events",
			},
		})
		return false, err
	}

	logger.Add(ctx, "code_exists", exists)
	return exists, nil
}

// UpdatePartial performs a partial update on an event, only updating fields that are explicitly provided
// This is a pure data access method - it executes the update and returns success/failure
// The usecase layer is responsible for validation and fetching updated data if needed
func (er *eventRepository) UpdatePartial(ctx context.Context, code string, updates *models.UpdateEventRequest) error {
	logger.AddProcess(ctx, "db_operation", "event.update_partial")

	// Build update map from model
	updateMap, fieldCount := updates.ToUpdateMap()

	// If no fields to update, return early
	if fieldCount == 0 {
		logger.Add(ctx, "no_fields_to_update", true)
		return nil
	}

	logger.Add(ctx, "fields_to_update", fieldCount)

	// Execute update
	err := er.db(ctx).
		Model(&models.Event{}).
		Where("code = ?", code).
		Updates(updateMap).
		Error

	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_UPDATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "events",
			},
		})
		return err
	}

	logger.Add(ctx, "update_succeeded", true)
	logger.Add(ctx, "fields_updated", fieldCount)

	return nil
}
