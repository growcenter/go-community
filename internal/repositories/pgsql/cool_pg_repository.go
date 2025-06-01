package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type CoolRepository interface {
	CheckByCode(ctx context.Context, code string) (dataExist bool, err error)
	GetOneByCode(ctx context.Context, code string) (cool models.Cool, err error)
	GetNameByCode(ctx context.Context, code string) (cool models.Cool, err error)
	Create(ctx context.Context, cool *models.Cool) (err error)
	GetAllOptions(ctx context.Context) (cool []models.GetAllCoolOptionsDBOutput, err error)
	GetCoolMemberByCode(ctx context.Context, code string) (cool []models.GetCoolMembersByIdDBOutput, err error)
	GetOneByCommunityId(ctx context.Context, communityId string) (cool models.Cool, err error)
	GetCoolFacilitatorByCode(ctx context.Context, code string) (facilitators []models.GetCoolMembersByIdDBOutput, err error)
	GetAllMembersByCode(ctx context.Context, code string) (members []models.GetCoolMembersByIdDBOutput, err error)
}

type coolRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewCoolRepository(db *gorm.DB, trx TransactionRepository) CoolRepository {
	return &coolRepository{db: db, trx: trx}
}

func (clr *coolRepository) CheckByCode(ctx context.Context, code string) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = clr.db.Raw(queryCheckCoolByCode, code).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}

func (clr *coolRepository) GetOneByCode(ctx context.Context, code string) (cool models.Cool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var cl models.Cool
	err = clr.db.Where("id = ?", code).Find(&cl).Error

	return cl, err
}

func (clr *coolRepository) GetNameByCode(ctx context.Context, code string) (cool models.Cool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var cl models.Cool
	err = clr.db.Raw(queryGetNameByCode, code).Scan(&cl).Error

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

func (clr *coolRepository) GetCoolMemberByCode(ctx context.Context, code string) (cool []models.GetCoolMembersByIdDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var cl []models.GetCoolMembersByIdDBOutput
	err = clr.db.Raw(queryGetCoolMemberByCode, code).Scan(&cl).Error

	return cl, err
}

func (clr *coolRepository) GetOneByCommunityId(ctx context.Context, communityId string) (cool models.Cool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = clr.db.Raw(queryGetCoolByCommunityId, communityId).Scan(&cool).Error

	return cool, err
}

func (clr *coolRepository) GetCoolFacilitatorByCode(ctx context.Context, code string) (facilitators []models.GetCoolMembersByIdDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = clr.db.Raw(queryGetCoolFacilitatorByCode, code).Scan(&facilitators).Error

	return facilitators, err
}

func (clr *coolRepository) GetAllMembersByCode(ctx context.Context, code string) (members []models.GetCoolMembersByIdDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = clr.db.Raw(queryGetAllMembersByCoolCode, code, code).Scan(&members).Error

	return members, err
}
