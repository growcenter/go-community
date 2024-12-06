package usecases

import (
	"context"
	"go-community/internal/common"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"strings"
)

type EventRegistrationRecordUsecase interface {
	Create(ctx context.Context, request *models.CreateEventRegistrationRecordRequest, communityId string, roles []string) (response *models.CreateEventRegistrationRecordResponse, err error)
	GetAll(ctx context.Context) (userTypes []models.UserType, err error)
}

type eventRegistrationRecordUsecase struct {
	r pgsql.PostgreRepositories
}

func NewEventRegistrationRecordUsecase(r pgsql.PostgreRepositories) *eventRegistrationRecordUsecase {
	return &eventRegistrationRecordUsecase{
		r: r,
	}
}

func (erru *eventRegistrationRecordUsecase) Create(ctx context.Context, request *models.CreateEventRegistrationRecordRequest, roles []string) (response *models.CreateEventRegistrationRecordResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	event, err := erru.r.Event.GetOneByCode(ctx, common.StringTrimSpaceAndUpper(request.EventCode))
	if err != nil {
		return nil, err
	}

	isAllowedRoles := common.CheckOneDataInList(event.EventAllowedRoles, roles)
	switch {
	case event.EventCode == "":
		err = models.ErrorDataNotFound
		return
	case event.EventStatus == "inactive":
		err = models.ErrorEventNotValid
		return
	case common.Now().Before(event.EventRegisterStartAt.In(common.GetLocation())):
		err = models.ErrorCannotRegisterYet
		return
	case common.Now().After(event.EventRegisterEndAt.In(common.GetLocation())):
		err = models.ErrorRegistrationTimeDisabled
		return
	case !isAllowedRoles:
		err = models.ErrorForbiddenRole
		return
	}

	instance, err := erru.r.EventInstance.GetByCode(ctx, common.StringTrimSpaceAndUpper(request.InstanceCode))
	if err != nil {
		return nil, err
	}

	countTotalRegistrants := 1 + len(request.Registrants)
	switch {
	case instance.ID == 0:
		err = models.ErrorDataNotFound
		return
	case instance.EventCode != common.StringTrimSpaceAndUpper(request.EventCode):
		err = models.ErrorEventNotValid
		return
	case instance.EventCode != event.EventCode:
		err = models.ErrorEventNotValid
		return
	case strings.ToLower(instance.Status) == "inactive":
		err = models.ErrorRegistrationTimeDisabled
		return
	case !instance.IsRequired:
		err = models.ErrorNoRegistrationNeeded
		return
	case countTotalRegistrants > instance.MaxPerTransaction:
		err = models.ErrorExceedMaxSeating
		return
	case (instance.TotalSeats - instance.BookedSeats) <= 0:
		err = models.ErrorRegisterQuotaNotAvailable
		return
	case ((instance.TotalSeats - instance.BookedSeats - countTotalRegistrants) <= 0) && instance.IsRequired == true && event.EventIsRecurring == false:
		err = models.ErrorRegisterQuotaNotAvailable
		return
	}

	if request.IsUsingQR {

	}

	return nil, nil
}
