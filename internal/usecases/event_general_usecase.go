package usecases

import (
	"context"
	"go-community/internal/common"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
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

	for _, event := range data {
		switch {
		case common.Now().Before(event.OpenRegistration.In(common.GetLocation())) && event.Status != "closed":
			event.Status = "closed"
			if err = egu.egr.BulkUpdate(ctx, event); err != nil {
				return
			}
		case common.Now().After(event.ClosedRegistration.In(common.GetLocation())) && event.Status != "closed":
			event.Status = "closed"
			if err = egu.egr.BulkUpdate(ctx, event); err != nil {
				return
			}
		default:
			event.Status = "active"
		}
	}

	new, err := egu.egr.GetAll(ctx)
	if err != nil {
		return
	}

	detail := models.GetGeneralEventDetailResponse{
		Type:        models.TYPE_DETAIL,
		CurrentTime: common.Now(),
		IsUserValid: true,
	}

	return detail, new, nil
}
