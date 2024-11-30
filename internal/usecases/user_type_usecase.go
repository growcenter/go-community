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
	utr pgsql.UserTypeRepository
	rr  pgsql.RoleRepository
}

func NewUserTypeUsecase(utr pgsql.UserTypeRepository, rr pgsql.RoleRepository) *userTypeUsecase {
	return &userTypeUsecase{
		utr: utr,
		rr:  rr,
	}
}

func (utu *userTypeUsecase) Create(ctx context.Context, request *models.CreateUserTypeRequest) (userType *models.UserType, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	countRole, err := utu.rr.CheckMultiple(ctx, request.Roles)
	if err != nil {
		return nil, err
	}

	if int(countRole) != len(request.Roles) {
		return nil, models.ErrorDataNotFound
	}

	exist, err := utu.utr.Check(ctx, request.UserType)
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
	}

	if err := utu.utr.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

func (utu *userTypeUsecase) GetAll(ctx context.Context) (userTypes []models.UserType, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	data, err := utu.utr.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}
