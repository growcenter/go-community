package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)
type EventCommunityRequestRepository interface {
	Create(ctx context.Context, request *models.EventCommunityRequest) (err error)
	GetByID(ctx context.Context, id int) (request models.EventCommunityRequest, err error)
	GetAllByAccountNumber(ctx context.Context, accountNumber string) (requests []models.EventCommunityRequest, err error)
}
type eventCommunityRequestRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewEventCommunityRequestRepository(db *gorm.DB, trx TransactionRepository) EventCommunityRequestRepository {
	return &eventCommunityRequestRepository{db: db, trx: trx}
}

// Create - Insert a new community request into the database
func (r *eventCommunityRequestRepository) Create(ctx context.Context, request *models.EventCommunityRequest) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	// Start a transaction and insert the request
	return r.trx.Transaction(func(dtx *gorm.DB) error {
		return r.db.Create(request).Error
	})
}

// GetByID - Retrieve a community request by its ID
func (r *eventCommunityRequestRepository) GetByID(ctx context.Context, id int) (request models.EventCommunityRequest, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	// Find the request by ID
	err = r.db.Where("id = ?", id).First(&request).Error
	return request, err
}

// GetAllByCommunityNumber - Retrieve all community requests for a specific community number
func (r *eventCommunityRequestRepository) GetAllByAccountNumber(ctx context.Context, accountNumber string) (requests []models.EventCommunityRequest, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	// Find all community requests for the given community number
	err = r.db.Where("account_number = ?", accountNumber).Find(&requests).Error
	return requests, err
}
