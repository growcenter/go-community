package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type UserUsecase interface {
	Create(ctx context.Context, request *models.CreateUserCoolRequest) (user *models.User, err error)
}

type userUsecase struct {
	ur  pgsql.UserRepository
	cr  pgsql.CampusRepository
	ccr pgsql.CoolCategoryRepository
}

func NewUserUsecase(ur pgsql.UserRepository, cr pgsql.CampusRepository, ccr pgsql.CoolCategoryRepository) *userUsecase {
	return &userUsecase{
		ur: ur,
	}
}

func (uu *userUsecase) Create(ctx context.Context, request *models.CreateUserCoolRequest) (user *models.User, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	existEmail, err := uu.ur.GetByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	if existEmail.ID != 0 {
		return nil, models.ErrorAlreadyExist
	}

	existPhone, err := uu.ur.GetByPhoneNumber(ctx, request.PhoneNumber)
	if err != nil {
		return nil, err
	}

	if existPhone.ID != 0 {
		return nil, models.ErrorAlreadyExist
	}

	campus, err := uu.cr.GetByCode(ctx, request.CampusCode)
	if err != nil {
		return nil, err
	}

	if campus.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	coolCategory, err := uu.ccr.GetByCode(ctx, request.CoolCategoryCode)
	if err != nil {
		return nil, err
	}

	if coolCategory.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	return

	// input := models.CoolCategory{
	// 	Code:     request,
	// 	Name:     request.Name,
	// 	AgeStart: request.AgeStart,
	// 	AgeEnd:   request.AgeEnd,
	// 	Status:   request.Status,
	// }

	// if err := ccu.ccr.Create(ctx, &input); err != nil {
	// 	return nil, err
	// }

	// return &input, nil
}
