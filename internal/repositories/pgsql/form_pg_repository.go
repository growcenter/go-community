package pgsql

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormRepository interface {
	Create(ctx context.Context, form *models.Form) error
	GetByCode(ctx context.Context, code uuid.UUID) (form models.Form, err error)
	GetByCodes(ctx context.Context, codes []uuid.UUID) (forms []models.Form, err error)
}

type formRepository struct {
	*BaseRepository
}

// Compile-time interface compliance check
var _ FormRepository = (*formRepository)(nil)

func NewFormRepository(db *gorm.DB) FormRepository {
	return &formRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new form
// Returns error if the form could not be created
func (r *formRepository) Create(ctx context.Context, form *models.Form) error {
	logger.Add(ctx, "db_operation", "form.create")
	err := r.db(ctx).Create(form).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_CREATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "forms",
			},
		})
	}
	return err
}

// GetByCode retrieves a form by its code
// Returns the form if found, or an error if not found or on database error
func (r *formRepository) GetByCode(ctx context.Context, code uuid.UUID) (form models.Form, err error) {
	logger.Add(ctx, map[string]any{
		"db_operation": "form.get_by_code",
		"lookup_code":  code,
	})

	err = r.db(ctx).Where("code = ?", code).First(&form).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "forms",
			},
		})
		return form, err
	}

	logger.Add(ctx, "form_found", true)
	return form, nil
}

// GetByCodes retrieves multiple forms by their codes
// Returns a list of forms or an error
func (r *formRepository) GetByCodes(ctx context.Context, codes []uuid.UUID) (forms []models.Form, err error) {
	if len(codes) == 0 {
		return []models.Form{}, nil
	}

	logger.Add(ctx, map[string]any{
		"db_operation": "form.get_by_codes",
		"code_count":   len(codes),
	})

	err = r.db(ctx).Where("code IN ?", codes).Find(&forms).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "forms",
			},
		})
		return nil, err
	}
	return forms, nil
}
