package usecases

import (
	"context"
	"go-community/internal/config"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type EventAttendanceUsecase interface {
	Create(ctx context.Context, request models.CreateEventAttendanceRequest) (*models.Attendee, error)
}

type eventAttendanceUsecase struct {
	cfg *config.Configuration
	r   pgsql.PostgreRepositories
}

func NewEventAttendanceUsecase(cfg config.Configuration, r pgsql.PostgreRepositories) *eventAttendanceUsecase {
	return &eventAttendanceUsecase{
		cfg: &cfg,
		r:   r,
	}
}

// func (u *eventAttendanceUsecase) Create(ctx context.Context, request models.CreateEventAttendanceRequest) (*models.Attendee, error) {
// 	// Get event registration
// 	registration, err := u.r.EventRegistration.GetByCode(ctx, request.RegistrationCode)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create event attendance
// 	attendance := &models.Attendee{
// 		Code:             uuid.New(),
// 		RegistrationCode: registration.Code,
// 		Role:             "attendee",
// 		Name:             request.Name,
// 		IsVerified:       false,
// 	}

// 	if err := u.r.EventAttendance.Create(ctx, attendance); err != nil {
// 		return nil, err
// 	}

// 	return attendance, nil
// }
