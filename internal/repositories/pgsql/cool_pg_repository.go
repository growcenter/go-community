package pgsql

import (
	"context"
	"go-community/internal/models"
	"gorm.io/gorm"
)

type CoolRepository interface {
	CheckById(ctx context.Context, id int) (dataExist bool, err error)
	GetOneById(ctx context.Context, id int) (cool models.Cool, err error)
	GetNameById(ctx context.Context, id int) (cool models.Cool, err error)
	Create(ctx context.Context, cool *models.Cool) (err error)
	GetAllOptions(ctx context.Context) (cool []models.GetAllCoolOptionsDBOutput, err error)
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

func (clr *coolRepository) GetOneById(ctx context.Context, id int) (cool models.Cool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var cl models.Cool
	err = clr.db.Where("id = ?", id).Find(&cl).Error

	return cl, err
}

func (clr *coolRepository) GetNameById(ctx context.Context, id int) (cool models.Cool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var cl models.Cool
	err = clr.db.Raw(queryGetNameById, id).Scan(&cl).Error

	return cl, err
}

func (clr *coolRepository) Create(ctx context.Context, cool *models.Cool) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return clr.db.Create(&cool).Error
}

func (clr *coolRepository) GetAllOptions(ctx context.Context) (cool []models.GetAllCoolOptionsDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var cl []models.GetAllCoolOptionsDBOutput
	err = clr.db.Raw(queryGetCoolsOptions).Scan(&cl).Error

	return cl, err
}
