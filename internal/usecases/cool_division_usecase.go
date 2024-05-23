package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type CoolCategoryUsecase interface {
	Create(ctx context.Context, request *models.CreateCoolCategoryRequest) (coolDivision *models.CoolCategory, err error)
}

type coolCategoryUsecase struct {
	cdr	pgsql.CoolCategoryRepository
}

func NewCoolDivisionUsecase(cdr pgsql.CoolCategoryRepository) *coolCategoryUsecase {
	return &coolCategoryUsecase{
		cdr: cdr,
	}
}

func (cdu *coolCategoryUsecase) Create(ctx context.Context, request *models.CreateCoolCategoryRequest) (coolDivision *models.CoolCategory, err error) {
	defer func() {
        LogService(ctx, err)
    }()

	exist, err := cdu.cdr.GetByCode(ctx, request.Code)
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

	if err := cdu.cdr.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}