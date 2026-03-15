package pgsql

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/logger"

	"gorm.io/gorm"
)

type FormAssociationRepository interface {
	Create(ctx context.Context, association *models.FormAssociation) error
	GetByFormCode(ctx context.Context, formCode string) ([]models.FormAssociation, error)
	GetByEntity(ctx context.Context, entityCode string, entityType string) ([]models.FormAssociation, error)
	Delete(ctx context.Context, formCode, entityCode string) error
	GetDummy(ctx context.Context, eventCode string, eventSessionCode string) ([]map[string]any, error)
}

type formAssociationRepository struct {
	*BaseRepository
}

var _ FormAssociationRepository = (*formAssociationRepository)(nil)

func NewFormAssociationRepository(db *gorm.DB) FormAssociationRepository {
	return &formAssociationRepository{BaseRepository: NewBaseRepository(db)}
}

// Create creates a new form association.
// Participates in any enclosing Atomic transaction via context.
func (r *formAssociationRepository) Create(ctx context.Context, association *models.FormAssociation) (err error) {
	logger.AddProcess(ctx, "db_operation", "form_association.create")
	err = r.db(ctx).Create(association).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_ASSOCIATION_CREATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_associations",
			},
		})
	}
	return err
}

// GetByFormCode retrieves all associations for a given form code.
func (r *formAssociationRepository) GetByFormCode(ctx context.Context, formCode string) (associations []models.FormAssociation, err error) {
	logger.AddProcess(ctx, "db_operation", "form_association.get_by_form_code")
	logger.Add(ctx, "form_code", formCode)

	err = r.db(ctx).Where("form_code = ?", formCode).Find(&associations).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_ASSOCIATION_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_associations",
			},
		})
		return nil, err
	}
	return associations, nil
}

// GetByEntity retrieves all associations for a given entity.
func (r *formAssociationRepository) GetByEntity(ctx context.Context, entityType string, entityCode string) (associations []models.FormAssociation, err error) {
	logger.AddProcess(ctx, "db_operation", "form_association.get_by_entity")
	logger.Add(ctx, map[string]any{
		"entity_code": entityCode,
		"entity_type": entityType,
	})

	err = r.db(ctx).Where("entity_code = ? AND entity_type = ?", entityCode, entityType).Find(&associations).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_ASSOCIATION_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_associations",
			},
		})
		return nil, err
	}
	return associations, nil
}

// Delete removes a specific association between a form and an entity.
func (r *formAssociationRepository) Delete(ctx context.Context, formCode, entityCode string) error {
	logger.AddProcess(ctx, "db_operation", "form_association.delete")
	logger.Add(ctx, map[string]any{
		"form_code":   formCode,
		"entity_code": entityCode,
	})

	err := r.db(ctx).
		Where("form_code = ? AND entity_code = ?", formCode, entityCode).
		Delete(&models.FormAssociation{}).Error

	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_ASSOCIATION_DELETE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_associations",
			},
		})
		return err
	}
	return nil
}

// GetDummy is a temporary function to get all forms, form associations, and form questions
// for event and event_session in one query.
func (r *formAssociationRepository) GetDummy(ctx context.Context, eventCode string, eventSessionCode string) ([]map[string]any, error) {
	logger.AddProcess(ctx, "db_operation", "form_association.get_dummy")

	var results []map[string]any

	query := r.db(ctx).Table("form_associations").
		Select(`
			form_associations.form_code, 
			form_associations.entity_type, 
			form_associations.entity_code, 
			forms.name as form_name, 
			forms.form_type as form_type,
			form_questions.code as question_code, 
			form_questions.text as question_text, 
			form_questions.category as question_category
		`).
		Joins("LEFT JOIN forms ON forms.code = form_associations.form_code").
		Joins("LEFT JOIN form_questions ON form_questions.form_code = forms.code")

	// Filter by the specific event code OR the specific event session code
	query = query.Where(
		"(form_associations.entity_type = ? AND form_associations.entity_code = ?) OR (form_associations.entity_type = ? AND form_associations.entity_code = ?)",
		"event", eventCode,
		"event_session", eventSessionCode,
	)

	err := query.Find(&results).Error

	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_ASSOCIATION_DUMMY_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_associations",
			},
		})
		return nil, err
	}

	logger.Add(ctx, "dummy_records_found", len(results))
	return results, nil
}
