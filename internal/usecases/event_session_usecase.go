package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type EventSessionUsecase interface {
	GetAllByEventCode(ctx context.Context, eventCode string) (eventSessions []models.EventSession, err error)
}

type eventSessionUsecase struct {
	esr pgsql.EventSessionRepository
}

func NewEventSessionUsecase(esr pgsql.EventSessionRepository) *eventSessionUsecase {
	return &eventSessionUsecase{
		esr: esr,
	}
}

func (esu *eventSessionUsecase) GetAllByEventCode(ctx context.Context, eventCode string) (eventSessions []models.EventSession, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	data, err := esu.esr.GetAllByEventCode(ctx, eventCode)
	if err != nil {
		return
	}

	if len(data) == 0 {
		err = models.ErrorDataNotFound
		return
	}

	// currentTime := time.Now()
	// fmt.Printf("======= TIME NOW: %s =======", currentTime)
	// for _, event := range data {
	// 	if currentTime.After(event.OpenRegistration) && currentTime.Before(event.ClosedRegistration) {
	// 		event.Status = "active"
	// 	} else {
	// 		event.Status = "closed"
	// 	}

	// 	if err = egu.egr.BulkUpdate(ctx, event); err != nil {
	// 		fmt.Println("=============error here")
	// 		return
	// 	}
	// }

	// new, err := egu.egr.GetAll(ctx)
	// if err != nil {
	// 	return
	// }

	// detail := models.GetEventSessionsDetailResponse{
	// 	Type:        models.TYPE_DETAIL,
	// 	CurrentTime: currentTime,
	// 	IsUserValid: true,
	// }

	return data, nil
}
