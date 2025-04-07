package pgsql

import (
	"context"
	"go-community/internal/models"
	"time"

	"gorm.io/gorm"
)

type FeatureFlagRepository interface {
	GetAll(ctx context.Context) (flags []models.FeatureFlag, err error)
	GetByKey(ctx context.Context, key string) (flag models.FeatureFlag, err error)
	Create(ctx context.Context, flag models.FeatureFlag) (err error)
	Update(ctx context.Context, flag *models.FeatureFlag) error
	Toggle(ctx context.Context, key string, enabled bool) error
	Delete(ctx context.Context, key string) error
}

type featureFlagRepository struct {
	db *gorm.DB
}

func NewFeatureFlagRepository(db *gorm.DB) FeatureFlagRepository {
	return &featureFlagRepository{db: db}
}

func (ffr *featureFlagRepository) GetAll(ctx context.Context) (flags []models.FeatureFlag, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var ff []models.FeatureFlag
	err = ffr.db.Find(&ff).Error

	return ff, err
}

// GetByKey retrieves a feature flag by its key
func (ffr *featureFlagRepository) GetByKey(ctx context.Context, key string) (flag models.FeatureFlag, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var ff models.FeatureFlag
	err = ffr.db.Where("key = ?", key).Find(&ff).Error

	return ff, err
}

// Create inserts a new feature flag
func (ffr *featureFlagRepository) Create(ctx context.Context, flag models.FeatureFlag) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return ffr.db.Create(&flag).Error
}

// Update updates an existing feature flag
func (ffr *featureFlagRepository) Update(ctx context.Context, flag *models.FeatureFlag) error {
	return ffr.db.WithContext(ctx).Model(&models.FeatureFlag{}).
		Where("id = ?", flag.ID).
		Updates(map[string]interface{}{
			"name":        flag.Name,
			"description": flag.Description,
			"enabled":     flag.Enabled,
			"rules":       flag.Rules,
			"updated_at":  time.Now(),
		}).Error
}

// ToggleFlag enables or disables a feature flag
func (ffr *featureFlagRepository) Toggle(ctx context.Context, key string, enabled bool) error {
	return ffr.db.WithContext(ctx).Model(&models.FeatureFlag{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"enabled":    enabled,
			"updated_at": time.Now(),
		}).Error

	//id, _ := strconv.Atoi(key)
	//
	//return ffr.db.WithContext(ctx).Model(&models.FeatureFlag{}).
	//	Where("id = ?", id).
	//	Updates(map[string]interface{}{
	//		"enabled":    enabled,
	//		"updated_at": time.Now(),
	//	}).Error
}

// Delete removes a feature flag by ID
func (ffr *featureFlagRepository) Delete(ctx context.Context, key string) error {
	return ffr.db.WithContext(ctx).Where("key = ?", key).Delete(&models.FeatureFlag{}).Error
}
