package pgsql

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/logger"

	"gorm.io/gorm"
)

type FormQuestionRepository interface {
	Create(ctx context.Context, formQuestion *models.FormQuestion) error
	BulkCreate(ctx context.Context, formQuestions *[]models.FormQuestion) error
	GetByFormCode(ctx context.Context, formCode string) (formQuestions []models.FormQuestion, err error)
	GetByFormCodes(ctx context.Context, formCodes []string) (formQuestions []models.FormQuestion, err error)
	GetByAssociationEntity(ctx context.Context, entities []models.FormQuestionEntityFilter) (formQuestions []models.FormQuestion, err error)
}

type formQuestionRepository struct {
	*BaseRepository
}

var _ FormQuestionRepository = (*formQuestionRepository)(nil)

func NewFormQuestionRepository(db *gorm.DB) FormQuestionRepository {
	return &formQuestionRepository{BaseRepository: NewBaseRepository(db)}
}

// Create creates a new form question
// Returns error if the question could not be created
func (r *formQuestionRepository) Create(ctx context.Context, formQuestion *models.FormQuestion) error {
	logger.Add(ctx, "db_operation", "form_question.create")
	err := r.db(ctx).Create(formQuestion).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_QUESTION_CREATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_questions",
			},
		})
	}
	return err
}

// BulkCreate inserts multiple form questions in a single database operation
// Uses GORM's CreateInBatches for efficient bulk insertion
// Returns error if any question could not be created
func (r *formQuestionRepository) BulkCreate(ctx context.Context, formQuestions *[]models.FormQuestion) error {
	if formQuestions == nil || len(*formQuestions) == 0 {
		return nil
	}

	logger.Add(ctx, map[string]any{
		"db_operation": "form_question.bulk_create",
		"record_count": len(*formQuestions),
	})

	err := r.db(ctx).CreateInBatches(formQuestions, 100).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_QUESTION_BULK_CREATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_questions",
			},
		})
	}
	return err
}

// GetByFormCode retrieves questions for a specific form
// Returns a list of form questions or an error
func (r *formQuestionRepository) GetByFormCode(ctx context.Context, formCode string) (formQuestions []models.FormQuestion, err error) {
	logger.Add(ctx, map[string]any{
		"db_operation": "form_question.get_by_form_code",
		"form_code":    formCode,
	})

	err = r.db(ctx).Where("form_code = ?", formCode).Find(&formQuestions).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_QUESTION_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_questions",
			},
		})
		return nil, err
	}
	return formQuestions, nil
}

// GetByFormCodes retrieves questions for multiple forms
// Returns a list of form questions or an error
func (r *formQuestionRepository) GetByFormCodes(ctx context.Context, formCodes []string) (formQuestions []models.FormQuestion, err error) {
	if len(formCodes) == 0 {
		return []models.FormQuestion{}, nil
	}

	logger.Add(ctx, map[string]any{
		"db_operation": "form_question.get_by_form_codes",
		"code_count":   len(formCodes),
	})

	err = r.db(ctx).Where("form_code IN ?", formCodes).Find(&formQuestions).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_QUESTION_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_questions",
			},
		})
		return nil, err
	}
	return formQuestions, nil
}

// GetByAssociationEntity retrieves questions associated with specific entities
// Uses a dynamic query builder to handle complex entity filtering
// Returns a list of form questions or an error
func (r *formQuestionRepository) GetByAssociationEntity(ctx context.Context, entities []models.FormQuestionEntityFilter) (formQuestions []models.FormQuestion, err error) {
	if len(entities) == 0 {
		return []models.FormQuestion{}, nil
	}

	logger.Add(ctx, map[string]any{
		"db_operation": "form_question.get_by_association_entity",
		"entity_count": len(entities),
	})

	query, args := BuildGetFormQuestionsByEntitiesQuery(entities)
	err = r.db(ctx).Raw(query, args...).Scan(&formQuestions).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_QUESTION_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_questions",
			},
		})
		return nil, err
	}
	return formQuestions, nil
}
