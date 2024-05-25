package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type CoolCategoryUsecase interface {
	Create(ctx context.Context, request *models.CreateCoolCategoryRequest) (coolCategory *models.CoolCategory, err error)
	GetAll(ctx context.Context) (coolCategories []*models.CoolCategory, err error)
}

type coolCategoryUsecase struct {
	ccr	pgsql.CoolCategoryRepository
}

func NewCoolDivisionUsecase(ccr pgsql.CoolCategoryRepository) *coolCategoryUsecase {
	return &coolCategoryUsecase{
		ccr: ccr,
	}
}

func (ccu *coolCategoryUsecase) Create(ctx context.Context, request *models.CreateCoolCategoryRequest) (coolCategories *models.CoolCategory, err error) {
	defer func() {
        LogService(ctx, err)
    }()

	exist, err := ccu.ccr.GetByCode(ctx, request.Code)
	if err != nil {
		return nil, err
	}

	if exist.ID != 0 {
		return nil, models.ErrorAlreadyExist
	}

	input := models.CoolCategory{
		Code: request.Code,
		Name: request.Name,
		AgeStart: request.AgeStart,
		AgeEnd: request.AgeEnd,
		Status: request.Status,
	}

	if err := ccu.ccr.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

func (ccu *coolCategoryUsecase) GetAll(ctx context.Context) (coolCategories []models.CoolCategory, err error) {
	defer func() {
        LogService(ctx, err)
    }()

	data, err := ccu.ccr.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}