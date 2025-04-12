package pgsql

import (
	"context"
	"go-community/internal/models"
	"gorm.io/gorm"
)

type ConfigRepository interface {
	GetByKey(ctx context.Context, key string) (config models.Config, err error)
}

type configRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepository{db: db}
}

// GetByKey retrieves a feature flag by its key
func (cdr *configRepository) GetByKey(ctx context.Context, key string) (config models.Config, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var c models.Config
	err = cdr.db.Where("key = ?", key).Find(&c).Error

	return c, err
}
