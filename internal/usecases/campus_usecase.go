package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"strings"
)

type CampusUsecase interface {
	Create(ctx context.Context, request *models.CreateCampusRequest) (user *models.Campus, err error)
	GetAll(ctx context.Context) (campus []models.Campus, err error)
}

type campusUsecase struct {
	cr	pgsql.CampusRepository
}

func NewCampusUsecase(cr pgsql.CampusRepository) *campusUsecase {
	return &campusUsecase{
		cr: cr,
	}
}

func (cu *campusUsecase) Create(ctx context.Context, request *models.CreateCampusRequest) (user *models.Campus, err error) {
	defer func() {
        LogService(ctx, err)
    }()

	code := strings.ToUpper(request.Code)
	exist, err := cu.cr.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if exist.ID != 0 {
		return nil, models.ErrorAlreadyExist
	}

	input := models.Campus{
		Code: code,
		Region: request.Region,
		Name: request.Name,
		Location: request.Location,
		Address: request.Address,
		Status: request.Status,
	}

	if err := cu.cr.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

func (cu *campusUsecase) GetAll(ctx context.Context) (campus []models.Campus, err error) {
	defer func() {
        LogService(ctx, err)
    }()

	data, err := cu.cr.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}