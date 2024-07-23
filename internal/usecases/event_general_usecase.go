package usecases

import (
	"context"
	"fmt"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"time"
)

type EventGeneralUsecase interface {
	GetAll(ctx context.Context) (details models.GetGeneralEventDetailResponse, eventGenerals []models.EventGeneral, err error)
}

type eventGeneralUsecase struct {
	egr pgsql.EventGeneralRepository
}

func NewEventGeneralUsecase(egr pgsql.EventGeneralRepository) *eventGeneralUsecase {
	return &eventGeneralUsecase{
		egr: egr,
	}
}

func (egu *eventGeneralUsecase) GetAll(ctx context.Context) (details models.GetGeneralEventDetailResponse, eventGenerals []models.EventGeneral, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	data, err := egu.egr.GetAll(ctx)
	if err != nil {
		return
	}

	currentTime := time.Now()
	fmt.Printf("======= TIME NOW: %s =======", currentTime)
	for _, event := range data {
		if currentTime.After(event.OpenRegistration) && currentTime.Before(event.ClosedRegistration) {
			event.Status = "active"
		} else {
			event.Status = "closed"
		}

		if err = egu.egr.BulkUpdate(ctx, event); err != nil {
			fmt.Println("=============error here")
			return
		}
	}

	new, err := egu.egr.GetAll(ctx)
	if err != nil {
		return
	}

	detail := models.GetGeneralEventDetailResponse{
		Type:        models.TYPE_DETAIL,
		CurrentTime: currentTime,
		IsUserValid: true,
	}

	return detail, new, nil
}
