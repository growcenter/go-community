package pgsql

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/logger"

	"gorm.io/gorm"
)

type FormAnswerRepository interface {
	BulkCreate(ctx context.Context, answers *[]models.FormAnswer) error
}

type formAnswerRepository struct {
	*BaseRepository
}

var _ FormAnswerRepository = (*formAnswerRepository)(nil)

func NewFormAnswerRepository(db *gorm.DB) FormAnswerRepository {
	return &formAnswerRepository{BaseRepository: NewBaseRepository(db)}
}

// BulkCreate inserts multiple form answers in a single database operation.
// Participates in any enclosing Atomic transaction via context.
func (r *formAnswerRepository) BulkCreate(ctx context.Context, answers *[]models.FormAnswer) error {
	if answers == nil || len(*answers) == 0 {
		return nil
	}

	logger.AddProcess(ctx, "db_operation", "form_answer.bulk_create")
	logger.Add(ctx, "record_count", len(*answers))

	err := r.db(ctx).CreateInBatches(answers, 100).Error
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:    "DatabaseError",
			Code:    "FORM_ANSWER_BULK_CREATE_FAILED",
			Message: err.Error(),
			Details: map[string]any{
				"db_table": "form_answers",
			},
		})
	}
	return err
}
