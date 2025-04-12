package pgsql

import (
	"context"
	"fmt"
	"go-community/internal/models"
	"go-community/internal/pkg/cursor"
	"gorm.io/gorm"
)

type CoolNewJoinerRepository interface {
	Create(ctx context.Context, question *models.CoolNewJoiner) (err error)
	GetAll(ctx context.Context, param models.GetAllCoolNewJoinerCursorParam) (output []models.GetCoolNewJoinerResponse, pagination *models.PaginationOutput, err error)
	GetById(ctx context.Context, id int) (output *models.CoolNewJoiner, err error)
	Update(ctx context.Context, question *models.CoolNewJoiner) (err error)
}

type coolNewJoinerRepository struct {
	db *gorm.DB
}

func NewCoolNewJoinerRepository(db *gorm.DB) CoolNewJoinerRepository {
	return &coolNewJoinerRepository{db: db}
}

func (cnjr *coolNewJoinerRepository) Create(ctx context.Context, question *models.CoolNewJoiner) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return cnjr.db.Create(&question).Error
}

func (cnjr *coolNewJoinerRepository) GetAll(ctx context.Context, param models.GetAllCoolNewJoinerCursorParam) (output []models.GetCoolNewJoinerResponse, pagination *models.PaginationOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	// Set default limit if none provided
	if param.Limit <= 0 {
		param.Limit = 10 // Default limit
	}

	queryList, paramList, err := BuildQueryGetAllCoolNewJoiner(param)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build list query: %w", err)
	}

	var records []models.GetCoolNewJoinerResponse
	if err := cnjr.db.Raw(queryList, paramList...).Scan(&records).Error; err != nil {
		return nil, nil, fmt.Errorf("query execution failed: %w", err)
	}

	queryCount, paramCount, err := BuildCountGetAllCoolNewJoiner(param)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build count query: %w", err)
	}

	var total int
	if err := cnjr.db.Raw(queryCount, paramCount...).Scan(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("count query execution failed: %w", err)
	}

	var next, prev string
	if len(records) > 0 {
		hasMore := len(records) > param.Limit
		isForward := param.Direction != "prev" && param.Cursor != ""
		isBackward := !isForward && param.Direction == "prev" && param.Cursor != ""

		if hasMore {
			records = records[:param.Limit] // Remove the extra record
		}

		if isBackward {
			for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
				records[i], records[j] = records[j], records[i]
			}
		}

		if isBackward || hasMore {
			lastRecord := records[len(records)-1]
			nextCursor := models.GetAllCoolNewJoinerCursor{
				CreatedAt: lastRecord.CreatedAt,
				ID:        lastRecord.ID,
			}
			next = cursor.EncryptCursorFromStruct(nextCursor)
		}

		if isForward || (hasMore && isBackward) {
			firstRecord := records[0]
			prevCursor := models.GetAllCoolNewJoinerCursor{
				CreatedAt: firstRecord.CreatedAt,
				ID:        firstRecord.ID,
			}
			prev = cursor.EncryptCursorFromStruct(prevCursor)
		}
	}

	pagination = &models.PaginationOutput{
		Next:  next,
		Prev:  prev,
		Total: total,
	}

	return records, pagination, nil
}

func (cnjr *coolNewJoinerRepository) GetById(ctx context.Context, id int) (output *models.CoolNewJoiner, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var record models.CoolNewJoiner
	if err := cnjr.db.Model(&models.CoolNewJoiner{}).Where("id = ?", id).First(&record).Error; err != nil {
		return nil, fmt.Errorf("failed to get cool new joiner by ID: %w", err)
	}

	return &record, nil
}

func (cnjr *coolNewJoinerRepository) Update(ctx context.Context, question *models.CoolNewJoiner) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return cnjr.db.Model(&models.CoolNewJoiner{}).Where("id = ?", question.ID).Updates(question).Error
}
