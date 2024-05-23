package pgsql

import (
	"context"

	"gorm.io/gorm"
)

type HealthRepository interface {
	Check(ctx context.Context) (err error)
}

type healthRepository struct {
	db *gorm.DB
}

func NewHealthRepository(db *gorm.DB) HealthRepository {
	return &healthRepository{db: db}
}

func (hr *healthRepository) Check(ctx context.Context) (err error) {
	defer func ()  {
		LogRepository(ctx, err)
	}()
	
	psql, err := hr.db.DB()
	if err != nil {
		return err
	}

	err = psql.PingContext(ctx)
	if err != nil {
		return err
	}

	return nil
}