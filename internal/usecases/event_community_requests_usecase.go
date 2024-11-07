package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type EventCommunityRequestUsecase interface {
	Create(ctx context.Context, request *models.CreateEventCommunityRequest) (*models.EventCommunityRequest, error)
	GetByID(ctx context.Context, id int) (*models.EventCommunityRequest, error)
	GetAllByAccountNumber(ctx context.Context, accountNumber string) ([]models.EventCommunityRequest, error)
}

type eventCommunityRequestUsecase struct {
	repo pgsql.EventCommunityRequestRepository
}

// NewEventCommunityRequestUsecase creates and returns a new eventCommunityRequestUsecase instance
func NewEventCommunityRequestUsecase(repo pgsql.EventCommunityRequestRepository) *eventCommunityRequestUsecase {
	return &eventCommunityRequestUsecase{
		repo: repo,
	}
}

// Create - Create a new event community request
func (ucr *eventCommunityRequestUsecase) Create(ctx context.Context, request *models.CreateEventCommunityRequest) (*models.EventCommunityRequest, error) {
	defer func() {
		LogService(ctx, nil) // You can improve logging here
	}()

	// Create the new request tied to the account_number
	newRequest := models.EventCommunityRequest{
		FullName:           request.FullName,
		RequestType:        request.RequestType,
		Email:              request.Email,
		PhoneNumber:        request.PhoneNumber,
		RequestInformation: request.RequestInformation,
		WantContacted:      request.WantContacted,
		AccountNumber:      request.AccountNumber, 
	}

	// Save the new request to the database
	if err := ucr.repo.Create(ctx, &newRequest); err != nil {
		return nil, err
	}

	return &newRequest, nil
}

// GetByID - Get a community request by ID
func (ucr *eventCommunityRequestUsecase) GetByID(ctx context.Context, id int) (*models.EventCommunityRequest, error) {
	defer func() {
		LogService(ctx, nil) // You can improve logging here
	}()

	// Get the request by ID
	request, err := ucr.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &request, nil
}

// GetAllByAccountNumber - Get all community requests for a specific account number
func (ucr *eventCommunityRequestUsecase) GetAllByAccountNumber(ctx context.Context, accountNumber string) ([]models.EventCommunityRequest, error) {
	defer func() {
		LogService(ctx, nil) // You can improve logging here
	}()

	// Get all requests for the given account number
	requests, err := ucr.repo.GetAllByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, err
	}

	return requests, nil
}
