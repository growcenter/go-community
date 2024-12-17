package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"strings"
)

type RoleUsecase interface {
	Create(ctx context.Context, request *models.CreateRoleRequest) (role *models.Role, err error)
	GetAll(ctx context.Context) (roles []models.Role, err error)
}

type roleUsecase struct {
	rr pgsql.RoleRepository
}

func NewRoleUsecase(rr pgsql.RoleRepository) *roleUsecase {
	return &roleUsecase{
		rr: rr,
	}
}

func (ru *roleUsecase) Create(ctx context.Context, request *models.CreateRoleRequest) (role *models.Role, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	exist, err := ru.rr.Check(ctx, strings.TrimSpace(strings.ToLower(request.Role)))
	if err != nil {
		return nil, err
	}

	if exist {
		return nil, models.ErrorAlreadyExist
	}

	input := models.Role{
		Role:        strings.TrimSpace(strings.ToLower(request.Role)),
		Description: request.Description,
	}

	if err := ru.rr.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

func (ru *roleUsecase) GetAll(ctx context.Context) (roles []models.Role, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	data, err := ru.rr.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}
