package pgsql

import (
	"context"
	"gorm.io/gorm"
)

type CoolRepository interface {
	CheckById(ctx context.Context, id int) (dataExist bool, err error)
}

type coolRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewCoolRepository(db *gorm.DB, trx TransactionRepository) CoolRepository {
	return &coolRepository{db: db, trx: trx}
}

func (clr *coolRepository) CheckById(ctx context.Context, id int) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = clr.db.Raw(queryCheckCoolById, id).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}
