package pgsql

import (
	"context"
	"github.com/lib/pq"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) (err error)
	GetByRole(ctx context.Context, role string) (roles models.Role, err error)
	GetAll(ctx context.Context) (roles []models.Role, err error)
	Check(ctx context.Context, role string) (dataExist bool, err error)
	CheckMultiple(ctx context.Context, roles []string) (count int64, err error)
	GetByArray(ctx context.Context, array []string) (roles []models.Role, err error)
}

type roleRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewRoleRepository(db *gorm.DB, trx TransactionRepository) RoleRepository {
	return &roleRepository{db: db, trx: trx}
}

func (rr *roleRepository) Create(ctx context.Context, role *models.Role) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return rr.trx.Transaction(func(dtx *gorm.DB) error {
		return rr.db.Create(&role).Error
	})
}

func (rr *roleRepository) GetByRole(ctx context.Context, role string) (roles models.Role, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var r models.Role
	err = rr.db.Where("role = ?", role).Find(&r).Error

	return r, err
}

func (rr *roleRepository) GetAll(ctx context.Context) (roles []models.Role, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var r []models.Role
	err = rr.db.Find(&r).Error

	return r, err
}

func (rr *roleRepository) Check(ctx context.Context, role string) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = rr.db.Raw(querySingleCheckRole, role).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}

func (rr *roleRepository) CheckMultiple(ctx context.Context, roles []string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = rr.db.Raw(queryMultipleCheckRole, pq.Array(roles)).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (rr *roleRepository) GetByArray(ctx context.Context, array []string) (roles []models.Role, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = rr.db.Raw(queryGetRolesByArray, pq.Array(array)).Scan(&roles).Error

	return roles, err
}
