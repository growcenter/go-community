package pgsql

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/logger"

	"gorm.io/gorm"
)

type EventInstanceRepository interface {
	Create(ctx context.Context, event *models.EventInstance) (err error)
	BulkCreate(ctx context.Context, events []models.EventInstance) error
	GetByCode(ctx context.Context, code string) (*models.EventInstance, error)
	GetByEventCode(ctx context.Context, code string) ([]models.EventInstance, error)
	CheckByCode(ctx context.Context, code string) (bool, error)
}

type eventInstanceRepository struct {
	*BaseRepository
}

func NewEventInstanceRepository(db *gorm.DB) EventInstanceRepository {
	return &eventInstanceRepository{BaseRepository: NewBaseRepository(db)}
}

// Create creates a new event instance
// Returns error if the event instance could not be created
func (er *eventInstanceRepository) Create(ctx context.Context, event *models.EventInstance) (err error) {
	logger.Add(ctx, "db_operation", "event_instance.create")
	err = er.db(ctx).Create(&event).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_INSTANCE_CREATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "event_instances",
			},
		})
	}
	return err
}

// BulkCreate inserts multiple event instances in a single database operation
// Uses GORM's CreateInBatches for efficient bulk insertion with automatic batching
// Returns error if any instance could not be created
func (er *eventInstanceRepository) BulkCreate(ctx context.Context, events []models.EventInstance) error {
	if len(events) == 0 {
		return nil
	}

	logger.Add(ctx, map[string]any{
		"db_operation": "event_instance.bulk_create",
		"record_count": len(events),
	})

	// Use CreateInBatches for better performance with large datasets
	// Batch size of 100 is a good balance between performance and memory
	err := er.db(ctx).CreateInBatches(events, 100).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_INSTANCE_BULK_CREATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "event_instances",
			},
		})
		return err
	}

	return nil
}

// GetByCode retrieves an event instance by its code
// Returns the event instance if found, or an error if not found or on database error
func (er *eventInstanceRepository) GetByCode(ctx context.Context, code string) (*models.EventInstance, error) {
	logger.Add(ctx, map[string]any{
		"db_operation": "event_instance.get_by_code",
		"lookup_code":  code,
	})

	var event models.EventInstance
	err := er.db(ctx).Where("code = ?", code).First(&event).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_INSTANCE_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "event_instances",
			},
		})
		return nil, err
	}

	logger.Add(ctx, "event_instance_found", true)
	return &event, nil
}

// GetByEventCode retrieves all event instances for a given event code
// Returns the event instances if found, or an error if not found or on database error
func (er *eventInstanceRepository) GetByEventCode(ctx context.Context, code string) ([]models.EventInstance, error) {
	logger.Add(ctx, map[string]any{
		"db_operation": "event_instance.get_by_event_code",
		"lookup_code":  code,
	})

	var events []models.EventInstance
	err := er.db(ctx).Where("event_code = ?", code).Find(&events).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_INSTANCE_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "event_instances",
			},
		})
		return nil, err
	}

	logger.Add(ctx, "event_instances_found", true)
	return events, nil
}

// CheckByCode checks if an instance code exists in the database
// Returns true if the code exists, false otherwise
// This is more efficient than GetByCode for uniqueness checks as it only counts
func (er *eventInstanceRepository) CheckByCode(ctx context.Context, code string) (bool, error) {
	logger.Add(ctx, map[string]any{
		"db_operation": "event_instance.check_by_code",
		"lookup_code":  code,
	})

	var count int64
	err := er.db(ctx).
		Model(&models.EventInstance{}).
		Where("code = ? AND deleted_at IS NULL", code).
		Count(&count).Error

	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "EVENT_INSTANCE_CHECK_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "event_instances",
			},
		})
		return false, err
	}

	exists := count > 0
	logger.Add(ctx, map[string]any{
		"code_exists": exists,
		"count":       count,
	})

	return exists, nil
}
