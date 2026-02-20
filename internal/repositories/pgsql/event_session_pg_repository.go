package pgsql

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/logger"

	"gorm.io/gorm"
)

type EventSessionRepository interface {
	// Create
	Create(ctx context.Context, event *models.EventSession) (err error)
	BulkCreate(ctx context.Context, events []models.EventSession) error

	// Get
	GetByCode(ctx context.Context, code string) (*models.EventSession, error)
	GetByEventCode(ctx context.Context, eventCode string) ([]models.EventSession, error)

	// Check
	CheckByCode(ctx context.Context, code string) (bool, error)
	CheckByEventCode(ctx context.Context, eventCode string) (bool, error)
}

type eventSessionRepository struct {
	*BaseRepository
}

// Compile-time interface compliance check
var _ EventSessionRepository = (*eventSessionRepository)(nil)

func NewEventSessionRepository(db *gorm.DB) EventSessionRepository {
	return &eventSessionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new event instance
// Returns error if the event instance could not be created
func (es *eventSessionRepository) Create(ctx context.Context, event *models.EventSession) (err error) {
	logger.Add(ctx, "db_operation", "event_session.create")

	err = es.db(ctx).Create(&event).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_SESSION_CREATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "event_sessions",
			},
		})
	}
	return err
}

// BulkCreate inserts multiple event instances in a single database operation
// Uses GORM's CreateInBatches for efficient bulk insertion with automatic batching
// Returns error if any instance could not be created
func (es *eventSessionRepository) BulkCreate(ctx context.Context, events []models.EventSession) error {
	if len(events) == 0 {
		logger.Add(ctx, "bulk_create_skipped", "no_events")
		return nil
	}

	logger.Add(ctx, map[string]any{
		"db_operation": "event_session.bulk_create",
		"record_count": len(events),
	})

	// Use CreateInBatches for better performance with large datasets
	// Batch size of 100 is a good balance between performance and memory
	err := es.db(ctx).CreateInBatches(events, 100).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_SESSION_BULK_CREATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "event_sessions",
			},
		})
		return err
	}

	logger.Add(ctx, map[string]any{
		"bulk_create_succeeded": true,
		"records_created":       len(events),
	})

	return nil
}

// GetByCode retrieves an event instance by its code
// Returns the event instance if found, or an error if not found or on database error
func (es *eventSessionRepository) GetByCode(ctx context.Context, code string) (*models.EventSession, error) {
	logger.Add(ctx, map[string]any{
		"db_operation": "event_session.get_by_code",
		"lookup_code":  code,
	})

	var event models.EventSession
	err := es.db(ctx).Where("code = ?", code).First(&event).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_SESSION_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table":             "event_sessions",
				"event_instance_found": false,
			},
		})
		return nil, err
	}

	logger.Add(ctx, "event_session_found", true)
	return &event, nil
}

// GetByEventCode retrieves an event instance by its event code
// Returns the event instance if found, or an error if not found or on database error
func (es *eventSessionRepository) GetByEventCode(ctx context.Context, eventCode string) ([]models.EventSession, error) {
	logger.Add(ctx, map[string]any{
		"db_operation": "event_session.get_by_event_code",
		"lookup_code":  eventCode,
	})

	var events []models.EventSession
	err := es.db(ctx).Where("event_code = ?", eventCode).Find(&events).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_SESSION_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table":             "event_sessions",
				"event_instance_found": false,
			},
		})
		return nil, err
	}

	logger.Add(ctx, "event_session_found", true)
	return events, nil
}

// CheckByCode checks if a session code exists in the database
// Uses PostgreSQL EXISTS for optimal performance - stops at first match instead of counting all rows
// Returns true if the code exists, false otherwise
func (es *eventSessionRepository) CheckByCode(ctx context.Context, code string) (bool, error) {
	logger.Add(ctx, map[string]any{
		"db_operation": "event_session.check_by_code",
		"lookup_code":  code,
	})

	var exists bool
	err := es.db(ctx).
		Raw(QueryCheckSessionByCode, code).
		Scan(&exists).Error

	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_SESSION_CHECK_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "event_sessions",
			},
		})
		return false, err
	}

	logger.Add(ctx, "code_exists", exists)
	return exists, nil
}

// CheckByEventCode checks if a session code exists in the database
// Uses PostgreSQL EXISTS for optimal performance - stops at first match instead of counting all rows
// Returns true if the code exists, false otherwise
func (es *eventSessionRepository) CheckByEventCode(ctx context.Context, eventCode string) (bool, error) {
	logger.Add(ctx, map[string]any{
		"db_operation": "event_session.check_by_event_code",
		"lookup_code":  eventCode,
	})

	var exists bool
	err := es.db(ctx).
		Raw(QueryCheckSessionByEventCode, eventCode).
		Scan(&exists).Error

	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_SESSION_CHECK_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "event_sessions",
			},
		})
		return false, err
	}

	logger.Add(ctx, "code_exists", exists)
	return exists, nil
}
