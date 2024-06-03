package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type LocationUsecase interface {
	Create(ctx context.Context, request *models.CreateLocationRequest) (location *models.Location, err error)
	GetAll(ctx context.Context) (locations []models.Location, err error)
	GetByCampusCode(ctx context.Context, campusCode string) (locations []models.Location, err error)
}

type locationUsecase struct {
	lr	pgsql.LocationRepository
	cr	pgsql.CampusRepository
}

func NewLocationUsecase(lr pgsql.LocationRepository, cr pgsql.CampusRepository) *locationUsecase {
	return &locationUsecase{
		lr: lr,
		cr: cr,
	}
}

func (lu *locationUsecase) Create(ctx context.Context, request *models.CreateLocationRequest) (location *models.Location, err error) {
	defer func() {
        LogService(ctx, err)
    }()	

	exist, err := lu.lr.GetByCode(ctx, request.Code)
	if err != nil {
		return nil, err
	}

	if exist.ID != 0 {
		return nil, models.ErrorAlreadyExist
	}

	campus, err := lu.cr.GetByCode(ctx, request.CampusCode)
	if err != nil {
		return nil, models.ErrorDataNotFound
	}

	if campus.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	input := models.Location {
		Code: request.Code,
		CampusCode: campus.Code,
		Name: request.Name,
		Region: campus.Region,
		Status: request.Status,
	}

	if err := lu.lr.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

func (lu *locationUsecase) GetAll(ctx context.Context) (locations []models.Location, err error) {
	defer func() {
        LogService(ctx, err)
    }()

	data, err := lu.lr.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (lu *locationUsecase) GetByCampusCode(ctx context.Context, campusCode string) (locations []models.Location, err error) {
	defer func() {
        LogService(ctx, err)
    }()

	data, err := lu.lr.GetByCampusCode(ctx, campusCode)
	if err != nil {
		return nil, models.ErrorDataNotFound
	}

	return data, nil
}