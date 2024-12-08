package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"strings"
)

type UserTypeUsecase interface {
	Create(ctx context.Context, request *models.CreateUserTypeRequest) (userType *models.UserType, err error)
	GetAll(ctx context.Context) (userTypes []models.UserType, err error)
}

type userTypeUsecase struct {
	r pgsql.PostgreRepositories
}

func NewUserTypeUsecase(r pgsql.PostgreRepositories) *userTypeUsecase {
	return &userTypeUsecase{
		r: r,
	}
}

func (utu *userTypeUsecase) Create(ctx context.Context, request *models.CreateUserTypeRequest) (userType *models.UserType, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	countRole, err := utu.r.Role.CheckMultiple(ctx, request.Roles)
	if err != nil {
		return nil, err
	}

	if int(countRole) != len(request.Roles) {
		return nil, models.ErrorDataNotFound
	}

	exist, err := utu.r.UserType.Check(ctx, request.UserType)
	if err != nil {
		return nil, err
	}

	if exist {
		return nil, models.ErrorAlreadyExist
	}

	input := models.UserType{
		Type:        strings.TrimSpace(strings.ToLower(request.UserType)),
		Name:        strings.TrimSpace(request.Name),
		Roles:       request.Roles,
		Description: request.Description,
		Category:    request.Category,
	}

	if err := utu.r.UserType.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

func (utu *userTypeUsecase) GetAll(ctx context.Context) (userTypes []models.UserType, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	data, err := utu.r.UserType.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}
