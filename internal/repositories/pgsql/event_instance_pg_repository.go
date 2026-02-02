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
	CheckByCode(ctx context.Context, code string) (bool, error)
}

type eventInstanceRepository struct {
	db *gorm.DB
}

func NewEventInstanceRepository(db *gorm.DB) EventInstanceRepository {
	return &eventInstanceRepository{db: db}
}

// Create creates a new event instance
// Returns error if the event instance could not be created
func (er *eventInstanceRepository) Create(ctx context.Context, event *models.EventInstance) (err error) {
	logger.EnrichContext(ctx, "db_operation", "event_instance.create")
	err = er.db.WithContext(ctx).Create(&event).Error
	if err != nil {
		logger.EnrichContextWith(ctx, map[string]any{
			"db_error":      true,
			"db_table":      "event_instances",
			"error_message": err.Error(),
		})
	}
	return err
}

// BulkCreate inserts multiple event instances in a single database operation
// Uses GORM's CreateInBatches for efficient bulk insertion with automatic batching
// Returns error if any instance could not be created
func (er *eventInstanceRepository) BulkCreate(ctx context.Context, events []models.EventInstance) error {
	if len(events) == 0 {
		logger.EnrichContext(ctx, "bulk_create_skipped", "no_events")
		return nil
	}

	logger.EnrichContextWith(ctx, map[string]any{
		"db_operation": "event_instance.bulk_create",
		"record_count": len(events),
	})

	// Use CreateInBatches for better performance with large datasets
	// Batch size of 100 is a good balance between performance and memory
	err := er.db.WithContext(ctx).CreateInBatches(events, 100).Error
	if err != nil {
		logger.EnrichContextWith(ctx, map[string]any{
			"db_error":      true,
			"db_table":      "event_instances",
			"error_message": err.Error(),
		})
		return err
	}

	logger.EnrichContextWith(ctx, map[string]any{
		"bulk_create_succeeded": true,
		"records_created":       len(events),
	})

	return nil
}

// GetByCode retrieves an event instance by its code
// Returns the event instance if found, or an error if not found or on database error
func (er *eventInstanceRepository) GetByCode(ctx context.Context, code string) (*models.EventInstance, error) {
	logger.EnrichContextWith(ctx, map[string]any{
		"db_operation": "event_instance.get_by_code",
		"lookup_code":  code,
	})

	var event models.EventInstance
	err := er.db.WithContext(ctx).Where("code = ?", code).First(&event).Error
	if err != nil {
		logger.EnrichContextWith(ctx, map[string]any{
			"db_error":             true,
			"db_table":             "event_instances",
			"event_instance_found": false,
			"error_message":        err.Error(),
		})
		return nil, err
	}

	logger.EnrichContext(ctx, "event_instance_found", true)
	return &event, nil
}

// CheckByCode checks if an instance code exists in the database
// Returns true if the code exists, false otherwise
// This is more efficient than GetByCode for uniqueness checks as it only counts
func (er *eventInstanceRepository) CheckByCode(ctx context.Context, code string) (bool, error) {
	logger.EnrichContextWith(ctx, map[string]any{
		"db_operation": "event_instance.check_by_code",
		"lookup_code":  code,
	})

	var count int64
	err := er.db.WithContext(ctx).
		Model(&models.EventInstance{}).
		Where("code = ? AND deleted_at IS NULL", code).
		Count(&count).Error

	if err != nil {
		logger.EnrichContextWith(ctx, map[string]any{
			"db_error":      true,
			"db_table":      "event_instances",
			"error_message": err.Error(),
		})
		return false, err
	}

	exists := count > 0
	logger.EnrichContextWith(ctx, map[string]any{
		"code_exists": exists,
		"count":       count,
	})

	return exists, nil
}
