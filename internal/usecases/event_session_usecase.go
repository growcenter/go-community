package usecases

import (
	"context"
	"go-community/internal/common"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type EventSessionUsecase interface {
	GetAllByEventCode(ctx context.Context, eventCode string) (details models.GetEventSessionsDetailResponse, eventSessions []models.EventSession, err error)
}

type eventSessionUsecase struct {
	esr pgsql.EventSessionRepository
	egr pgsql.EventGeneralRepository
}

func NewEventSessionUsecase(esr pgsql.EventSessionRepository, egr pgsql.EventGeneralRepository) *eventSessionUsecase {
	return &eventSessionUsecase{
		esr: esr,
		egr: egr,
	}
}

func (esu *eventSessionUsecase) GetAllByEventCode(ctx context.Context, eventCode string) (details models.GetEventSessionsDetailResponse, eventSessions []models.EventSession, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	event, err := esu.egr.GetByCode(ctx, eventCode)
	if err != nil {
		return
	}

	if event.ID == 0 {
		err = models.ErrorDataNotFound
		return
	}

	if common.Now().Before(event.OpenRegistration) {
		err = models.ErrorCannotRegisterYet
		return
	}

	if common.Now().After(event.ClosedRegistration) {
		err = models.ErrorRegistrationTimeDisabled
		return
	}

	data, err := esu.esr.GetAllByEventCode(ctx, eventCode)
	if err != nil {
		return
	}

	if len(data) == 0 {
		err = models.ErrorDataNotFound
		return
	}

	for _, session := range data {
		if session.AvailableSeats == 0 {
			session.Status = "full"
		}

		if err = esu.esr.BulkUpdate(ctx, session); err != nil {
			return
		}
	}

	new, err := esu.esr.GetAllByEventCode(ctx, eventCode)
	if err != nil {
		return
	}

	detail := models.GetEventSessionsDetailResponse{
		Type:        models.TYPE_DETAIL,
		EventCode:   event.Code,
		EventName:   event.Name,
		CurrentTime: common.Now(),
		IsUserValid: true,
	}

	return detail, new, nil
}
