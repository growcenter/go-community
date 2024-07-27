package usecases

import (
	"context"
	"go-community/internal/common"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"strings"

	"github.com/google/uuid"
)

type EventRegistrationUsecase interface {
	Create(ctx context.Context, request models.CreateEventRegistrationRequest) (eventRegistration models.CreateEventRegistrationResponse, err error)
	GetRegistered(ctx context.Context, registeredBy string) (eventRegistrations models.GetRegisteredResponse, err error)
}

type eventRegistrationUsecase struct {
	rer pgsql.EventRegistrationRepository
	egr pgsql.EventGeneralRepository
	esr pgsql.EventSessionRepository
	eur pgsql.EventUserRepository
}

func NewEventRegistrationUsecase(rer pgsql.EventRegistrationRepository, egr pgsql.EventGeneralRepository, esr pgsql.EventSessionRepository, eur pgsql.EventUserRepository) *eventRegistrationUsecase {
	return &eventRegistrationUsecase{
		rer: rer,
		egr: egr,
		esr: esr,
		eur: eur,
	}
}

func (eru *eventRegistrationUsecase) Create(ctx context.Context, request models.CreateEventRegistrationRequest) (eventRegistration models.CreateEventRegistrationResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	event, err := eru.egr.GetByCode(ctx, request.EventCode)
	if err != nil {
		return
	}

	switch {
	case event.ID == 0:
		err = models.ErrorDataNotFound
		return
	case event.Status == "closed":
		err = models.ErrorRegistrationTimeDisabled
		return
	case common.Now().Before(event.OpenRegistration.In(common.GetLocation())):
		err = models.ErrorCannotRegisterYet
	case common.Now().After(event.ClosedRegistration.In(common.GetLocation())):
		err = models.ErrorRegistrationTimeDisabled
	}

	// if event.ID == 0 {
	// 	err = models.ErrorDataNotFound
	// 	return
	// }

	// if event.Status == "closed" {
	// 	err = models.ErrorRegistrationTimeDisabled
	// 	return
	// }

	// if common.Now().Before(event.OpenRegistration.In(common.GetLocation())) {
	// 	err = models.ErrorCannotRegisterYet
	// 	return
	// }

	// if common.Now().After(event.ClosedRegistration.In(common.GetLocation())) {
	// 	err = models.ErrorRegistrationTimeDisabled
	// 	return
	// }

	session, err := eru.esr.GetByCode(ctx, request.SessionCode)
	if err != nil {
		return
	}

	countTotalRegister := 1 + len(request.Others)

	switch {
	case session.ID == 0:
		err = models.ErrorDataNotFound
		return
	case session.EventCode != event.Code:
		err = models.ErrorEventNotValid
		return
	case session.Status == "closed":
		err = models.ErrorRegistrationTimeDisabled
		return
	case session.Status == "full":
		err = models.ErrorRegisterQuotaNotAvailable
		return
	case countTotalRegister > session.MaxSeating:
		err = models.ErrorExceedMaxSeating
		return
	case session.AvailableSeats == 0:
		err = models.ErrorRegisterQuotaNotAvailable
		return
	case (session.AvailableSeats - countTotalRegister) < 0:
		err = models.ErrorRegisterQuotaNotAvailable
		return
	}

	// if session.ID == 0 {
	// 	err = models.ErrorDataNotFound
	// 	return
	// }

	// if session.EventCode != event.Code {
	// 	err = models.ErrorEventNotValid
	// 	return
	// }

	// if session.Status == "closed" {
	// 	err = models.ErrorRegistrationTimeDisabled
	// 	return
	// }

	// if countTotalRegister > session.MaxSeating {
	// 	err = models.ErrorExceedMaxSeating
	// 	return
	// }

	// if session.AvailableSeats == 0 {
	// 	err = models.ErrorRegisterQuotaNotAvailable
	// 	return
	// }

	// if (session.AvailableSeats - countTotalRegister) < 0 {
	// 	err = models.ErrorRegisterQuotaNotAvailable
	// 	return
	// }

	alreadyRegister, err := eru.rer.GetByRegisteredBy(ctx, strings.ToLower(request.Identifier))
	if err != nil {
		return
	}

	countTotalAlreadyRegister := len(alreadyRegister)
	if (countTotalRegister + countTotalAlreadyRegister) > session.MaxSeating {
		err = models.ErrorExceedMaxSeating
		return
	}

	user, err := eru.eur.GetByEmailPhone(ctx, request.Identifier)
	if err != nil {
		return
	}

	accountNumber := ""
	if user.ID > 0 {
		accountNumber = user.AccountNumber
		user.Address = request.Address

		if err = eru.eur.Update(ctx, &user); err != nil {
			return
		}
	}

	var register = make([]models.EventRegistration, 0, session.MaxSeating)

	mainInput := models.EventRegistration{
		Name:          strings.ToUpper(request.Name),
		Identifier:    strings.ToLower(request.Identifier),
		Address:       request.Address,
		AccountNumber: accountNumber,
		Code:          (uuid.New()).String(),
		EventCode:     event.Code,
		SessionCode:   session.Code,
		RegisteredBy:  strings.ToLower(request.Identifier),
		Status:        "registered",
	}

	register = append(register, mainInput)

	for _, other := range request.Others {
		otherInput := models.EventRegistration{
			Name:         strings.ToUpper(other.Name),
			Address:      other.Address,
			Code:         (uuid.New()).String(),
			EventCode:    event.Code,
			SessionCode:  session.Code,
			RegisteredBy: strings.ToLower(request.Identifier),
			Status:       "registered",
		}

		register = append(register, otherInput)
	}

	if err = eru.rer.BulkCreate(ctx, &register); err != nil {
		return
	}

	// Update Session Quota & Session
	session.AvailableSeats -= countTotalRegister
	if session.AvailableSeats == 0 {
		session.Status = "full"
	}

	if err = eru.esr.Update(ctx, session); err != nil {
		return
	}

	otherResponse := make([]models.CreateOtherEventRegistrationResponse, len(register))
	for i, p := range register {
		otherResponse[i] = models.CreateOtherEventRegistrationResponse{
			Type:    models.TYPE_EVENT_REGISTRATION,
			Name:    p.Name,
			Address: p.Address,
			Code:    p.Code,
		}
	}

	mainResponse := models.CreateEventRegistrationResponse{
		Type:          models.TYPE_EVENT_REGISTRATION,
		Name:          mainInput.Name,
		Identifier:    mainInput.Identifier,
		Address:       mainInput.Address,
		AccountNumber: mainInput.AccountNumber,
		Code:          mainInput.Code,
		EventCode:     event.Code,
		EventName:     event.Name,
		SessionCode:   session.Code,
		SessionName:   session.Name,
		IsValid:       true,
		Seats:         countTotalRegister,
		Status:        "registered",
		Others:        otherResponse[1:],
	}

	return mainResponse, nil
}

func (eru *eventRegistrationUsecase) GetRegistered(ctx context.Context, registeredBy string) (eventRegistrations []models.GetRegisteredResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	registers, err := eru.rer.GetSpecificByRegisteredBy(ctx, registeredBy)
	if err != nil {
		return
	}

	response := make([]models.GetRegisteredResponse, len(registers))
	for i, p := range registers {
		response[i] = models.GetRegisteredResponse{
			Type:          models.TYPE_EVENT_REGISTRATION,
			Name:          p.EventRegistration.Name,
			Identifier:    p.EventRegistration.Identifier,
			Address:       p.EventRegistration.Address,
			AccountNumber: p.EventRegistration.AccountNumber,
			Code:          p.EventRegistration.Code,
			EventCode:     p.EventRegistration.EventCode,
			EventName:     p.EventGeneral.Name,
			SessionCode:   p.EventRegistration.SessionCode,
			SessionName:   p.EventSession.Name,
			Status:        p.EventRegistration.Status,
		}
	}

	return response, nil
}
