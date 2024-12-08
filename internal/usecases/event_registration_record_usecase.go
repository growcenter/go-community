package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go-community/internal/common"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"time"
)

type EventRegistrationRecordUsecase interface {
	Create(ctx context.Context, request *models.CreateEventRegistrationRecordRequest, value *models.TokenValues) (response *models.CreateEventRegistrationRecordResponse, err error)
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

func (erru *eventRegistrationRecordUsecase) Create(ctx context.Context, request *models.CreateEventRegistrationRecordRequest, value *models.TokenValues) (response *models.CreateEventRegistrationRecordResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	if err = erru.validateCreate(ctx, request, value); err != nil {
		return nil, err
	}

	return erru.createAtomic(ctx, request, value)
}

func (erru *eventRegistrationRecordUsecase) createAtomic(ctx context.Context, request *models.CreateEventRegistrationRecordRequest, value *models.TokenValues) (response *models.CreateEventRegistrationRecordResponse, err error) {
	res := &models.CreateEventRegistrationRecordResponse{}

	var (
		registerStatus    string
		communityIdOrigin string
		verifiedAt        sql.NullTime
		updatedBy         string
		registerAt        time.Time
	)

	registerAt, _ = common.ParseStringToDatetime(time.RFC3339, request.RegisterAt, common.GetLocation())

	if request.IsPersonalQR {
		registerStatus = models.MapRegisterStatus[models.REGISTER_STATUS_SUCCESS]
		communityIdOrigin = request.CommunityId
		verifiedAt = sql.NullTime{
			Time:  registerAt,
			Valid: true,
		}
		updatedBy = "user"
	} else {
		registerStatus = models.MapRegisterStatus[models.REGISTER_STATUS_PENDING]
		communityIdOrigin = value.CommunityId
		verifiedAt = sql.NullTime{
			Valid: false,
		}
		updatedBy = ""
	}

	err = erru.r.Transaction.Atomic(ctx, func(ctx context.Context, r *pgsql.PostgreRepositories) error {
		countTotalRegistrants := 1 + len(request.Registrants)
		var register = make([]models.EventRegistrationRecord, 0, countTotalRegistrants)
		instance, err := r.EventInstance.GetSeatsNamesByCode(ctx, request.InstanceCode)
		if err != nil {
			return err
		}

		if instance == nil {
			return models.ErrorDataNotFound
		}

		main := models.EventRegistrationRecord{
			ID:                uuid.New(),
			Name:              common.StringTrimSpaceAndUpper(request.Name),
			Identifier:        request.Identifier,
			CommunityId:       request.CommunityId,
			EventCode:         request.EventCode,
			InstanceCode:      request.InstanceCode,
			IdentifierOrigin:  request.Identifier,
			CommunityIdOrigin: communityIdOrigin,
			Status:            registerStatus,
			RegisteredAt:      registerAt,
			VerifiedAt:        verifiedAt,
			UpdatedBy:         updatedBy,
		}

		register = append(register, main)

		for _, registrant := range request.Registrants {
			register = append(register, models.EventRegistrationRecord{
				ID:                uuid.New(),
				Name:              common.StringTrimSpaceAndUpper(registrant.Name),
				EventCode:         request.EventCode,
				InstanceCode:      request.InstanceCode,
				IdentifierOrigin:  request.Identifier,
				CommunityIdOrigin: communityIdOrigin,
				Status:            registerStatus,
				RegisteredAt:      registerAt,
			})
		}

		if err = r.EventRegistrationRecord.BulkCreate(ctx, &register); err != nil {
			return err
		}

		instance.BookedSeats += countTotalRegistrants

		if instance.TotalSeats != 0 {
			if instance.BookedSeats > instance.TotalSeats {
				return models.ErrorRegisterQuotaNotAvailable
			}

			if (instance.TotalRemainingSeats - instance.BookedSeats) <= 0 {
				return models.ErrorRegisterQuotaNotAvailable
			}
		}

		if err = r.EventInstance.UpdateBookedSeatsByCode(ctx, request.InstanceCode, instance); err != nil {
			return err
		}

		registrantRes := make([]models.CreateOtherEventRegistrationRecordResponse, len(register))
		for i, p := range register {
			registrantRes[i] = models.CreateOtherEventRegistrationRecordResponse{
				Type:   models.TYPE_EVENT_REGISTRATION,
				ID:     p.ID,
				Name:   p.Name,
				Status: p.Status,
			}
		}

		res = &models.CreateEventRegistrationRecordResponse{
			Type:             models.TYPE_EVENT_REGISTRATION,
			ID:               main.ID,
			Status:           registerStatus,
			Name:             main.Name,
			Identifier:       main.Identifier,
			CommunityID:      main.CommunityId,
			EventCode:        request.EventCode,
			EventTitle:       instance.EventTitle,
			InstanceCode:     request.InstanceCode,
			InstanceTitle:    instance.EventInstanceTitle,
			TotalRegistrants: countTotalRegistrants,
			RegisterAt:       registerAt,
			Registrants:      registrantRes[1:],
		}

		return nil
	})
	return res, err
}

func (erru *eventRegistrationRecordUsecase) validateCreate(ctx context.Context, request *models.CreateEventRegistrationRecordRequest, value *models.TokenValues) error {
	if request.EventCode != request.InstanceCode[:7] {
		return models.ErrorMismatchFields
	}

	if request.Identifier == "" && request.CommunityId == "" {
		return models.ErrorIdentifierCommunityIdEmpty
	}

	if request.IsPersonalQR {
		fmt.Println("sini")
		if request.CommunityId == "" {
			return models.ErrorInvalidInput
		}

		userExist, err := erru.r.User.GetByCommunityId(ctx, request.CommunityId)
		if err != nil {
			return err
		}

		if &userExist == nil {
			return models.ErrorDataNotFound
		}

	}

	countTotalRegistrants := 1 + len(request.Registrants)
	if request.IsPersonalQR && countTotalRegistrants > 1 {
		return models.ErrorQRForMoreThanOneRegister
	}

	event, err := erru.r.Event.GetOneByCode(ctx, request.EventCode)
	if err != nil {
		return err
	}

	eventAvailableStatus, err := models.DefineAvailabilityStatus(event)
	if err != nil {
		return err
	}

	registerAt, _ := common.ParseStringToDatetime(time.RFC3339, request.RegisterAt, common.GetLocation())
	fmt.Println("register : ", registerAt)
	fmt.Println("start : ", event.EventRegisterStartAt.In(common.GetLocation()))
	fmt.Println("end : ", event.EventRegisterEndAt.In(common.GetLocation()))

	switch {
	case event.EventCode == "" || event.EventStatus != models.MapStatus[models.STATUS_ACTIVE]:
		return models.ErrorDataNotFound
	case request.EventCode != event.EventCode:
		return models.ErrorEventNotValid
	//case common.Now().Before(event.EventRegisterStartAt.In(common.GetLocation())):
	//	return models.ErrorCannotRegisterYet
	//case common.Now().After(event.EventRegisterEndAt.In(common.GetLocation())):
	//	return models.ErrorRegistrationTimeDisabled
	case registerAt.Before(event.EventRegisterStartAt.In(common.GetLocation())):
		return models.ErrorCannotRegisterYet
	case registerAt.After(event.EventRegisterEndAt.In(common.GetLocation())):
		return models.ErrorRegistrationTimeDisabled
	//case request.IsPersonalQR && event.EventAllowedFor != "public":
	//	isAllowedRoles := common.CheckOneDataInList(event.EventAllowedRoles, value.Roles)
	//	isAllowedUsers := common.CheckOneDataInList(event.EventAllowedUsers, value.UserTypes)
	//	fmt.Println(isAllowedUsers, isAllowedRoles)
	//	if !isAllowedRoles && !isAllowedUsers {
	//		return models.ErrorForbiddenRole
	//	}
	case !request.IsPersonalQR && event.EventAllowedFor != "public":
		userExist, err := erru.r.User.GetByCommunityId(ctx, request.CommunityId)
		if err != nil {
			return err
		}

		if &userExist == nil {
			return models.ErrorDataNotFound
		}

		isAllowedRoles := common.CheckOneDataInList(event.EventAllowedRoles, userExist.Roles)
		isAllowedUsers := common.CheckOneDataInList(event.EventAllowedUsers, userExist.UserTypes)
		fmt.Println(isAllowedUsers, isAllowedRoles)
		if !isAllowedRoles && !isAllowedUsers {
			return models.ErrorForbiddenRole
		}
	case eventAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_UNAVAILABLE]:
		return models.ErrorEventNotAvailable
	case eventAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_FULL]:
		return models.ErrorRegisterQuotaNotAvailable
	case eventAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_SOON]:
		return models.ErrorCannotRegisterYet
	}

	instance, err := erru.r.EventInstance.GetOneByCode(ctx, request.InstanceCode, models.MapStatus[models.STATUS_ACTIVE])
	if err != nil {
		return err
	}

	instanceAvailableStatus, err := models.DefineAvailabilityStatus(instance)
	if err != nil {
		return err
	}

	switch {
	case instance.InstanceCode == "" || instance.InstanceStatus != models.MapStatus[models.STATUS_ACTIVE]:
		return models.ErrorDataNotFound
	case instance.InstanceEventCode != request.EventCode || instance.InstanceEventCode != event.EventCode:
		return models.ErrorEventNotValid
	case instance.InstanceRegisterFlow == models.MapRegisterFlow[models.REGISTER_FLOW_NONE]:
		return models.ErrorNoRegistrationNeeded
	case request.IsPersonalQR && instance.InstanceRegisterFlow == models.MapRegisterFlow[models.REGISTER_FLOW_EVENT]:
		return models.ErrorCannotUsePersonalQR
	//case instanceAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_UNAVAILABLE]:
	//	return models.ErrorEventNotAvailable
	case registerAt.After(event.EventRegisterEndAt.In(common.GetLocation())):
		return models.ErrorRegistrationTimeDisabled
	case ((instance.TotalRemainingSeats - countTotalRegistrants) <= 0) && instance.InstanceRegisterFlow != models.MapRegisterFlow[models.REGISTER_FLOW_NONE] && event.EventIsRecurring == false && instance.InstanceTotalSeats > 0:
		return models.ErrorRegisterQuotaNotAvailable
	case instanceAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_FULL]:
		//if event.EventAllowedFor != "private" {
		//	return models.ErrorRegisterQuotaNotAvailable
		//}

		return models.ErrorRegisterQuotaNotAvailable
	//case instanceAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_SOON]:
	//	return models.ErrorCannotRegisterYet
	case registerAt.Before(event.EventRegisterStartAt.In(common.GetLocation())):
		return models.ErrorCannotRegisterYet
	case instance.InstanceMaxPerTransaction > 0 && countTotalRegistrants > instance.InstanceMaxPerTransaction:
		fmt.Println("eror here")
		return models.ErrorExceedMaxSeating
	}

	switch {
	case instance.InstanceIsOnePerAccount:
		countRegistered, err := erru.r.EventRegistrationRecord.CountByCommunityIdOrigin(ctx, common.StringTrimSpaceAndLower(request.CommunityId))
		if err != nil {
			return err
		}
		if countRegistered > 0 {
			return models.ErrorEventCanOnlyRegisterOnce
		}
	case instance.InstanceIsOnePerTicket:
		if request.Identifier != "" && request.CommunityId == "" {
			identifierExist, err := erru.r.EventRegistrationRecord.CheckByIdentifier(ctx, common.StringTrimSpaceAndLower(request.Identifier))
			if err != nil {
				return err
			}
			if identifierExist {
				return models.ErrorAlreadyRegistered
			}
		} else if request.Identifier == "" && request.CommunityId != "" {
			communityIdExist, err := erru.r.EventRegistrationRecord.CheckByCommunityId(ctx, request.CommunityId)
			if err != nil {
				return err
			}
			if communityIdExist {
				return models.ErrorAlreadyRegistered
			}
		} else {
			return models.ErrorIdentifierCommunityIdEmpty
		}

		if len(request.Registrants) > 0 {
			for _, registrant := range request.Registrants {
				nameExist, err := erru.r.EventRegistrationRecord.CheckByName(ctx, common.StringTrimSpaceAndUpper(registrant.Name))
				if err != nil {
					return err
				}
				if nameExist {
					return models.ErrorAlreadyRegistered
				}
			}
		}
	case instance.InstanceIsOnePerTicket && instance.InstanceIsOnePerAccount:
		countRegistered, err := erru.r.EventRegistrationRecord.CountByCommunityIdOrigin(ctx, common.StringTrimSpaceAndLower(request.CommunityId))
		if err != nil {
			return err
		}
		if countRegistered > 0 {
			return models.ErrorEventCanOnlyRegisterOnce
		}
	}

	countRegistered, err := erru.r.EventRegistrationRecord.CountByIdentifierOriginAndStatus(ctx, common.StringTrimSpaceAndLower(request.Identifier), models.MapRegisterStatus[models.REGISTER_STATUS_PENDING])
	if err != nil {
		return err
	}
	if instance.InstanceMaxPerTransaction > 0 && ((int(countRegistered) + countTotalRegistrants) > instance.InstanceMaxPerTransaction) {
		return models.ErrorExceedMaxSeating
	}

	return nil
}
