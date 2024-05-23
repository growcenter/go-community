package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"strings"
)

type CampusUsecase interface {
	Create(ctx context.Context, request *models.CreateCampusRequest) (user *models.Campus, err error)
}

type campusUsecase struct {
	ur	pgsql.CampusRepository
}

func NewCampusUsecase(ur pgsql.CampusRepository) *campusUsecase {
	return &campusUsecase{
		ur: ur,
	}
}

func (cu *campusUsecase) Create(ctx context.Context, request *models.CreateCampusRequest) (user *models.Campus, err error) {
	defer func() {
        LogService(ctx, err)
    }()

	code := strings.ToUpper(request.Code)
	exist, err := cu.ur.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if exist.ID != 0 {
		return nil, models.ErrorUserNotFound
	}

	input := models.Campus{
		Code: code,
		Region: request.Region,
		Name: request.Name,
		Location: request.Location,
		Address: request.Address,
		Status: request.Status,
	}

	if err := cu.ur.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

// type CampusUsecase interface {
// 	Create(ctx context.Context, request *models.CreateCampusRequest) (err error)
// }

// type campus usecase

// var _ CampusUsecase = (*campus)(nil)

// func (cu *campus) Create(ctx context.Context, request *models.CreateCampusRequest) (err error) {
// 	defer func() {
// 		logService(ctx, err)
// 	}()

// 	return cu.u.postgreRepository.GetTransactionRepository().Transaction(func(tx *gorm.DB) error {
// 		exist, err := cu.u.postgreRepository.GetCampusRepository().GetByCode(ctx, request.Code)
// 		if err != nil {
// 			return err
// 		}

// 		if exist.ID == 0 {
// 			return models.ErrorUserNotFound
// 		}

// 		input := models.Campus{
// 			Code: strings.ToUpper(request.Code),
// 			Region: request.Region,
// 			Name: request.Name,
// 			Location: request.Location,
// 			Address: request.Address,
// 			Status: request.Status,
// 		}

// 		return cu.u.postgreRepository.GetCampusRepository().Create(ctx, &input)
// 	})
// }