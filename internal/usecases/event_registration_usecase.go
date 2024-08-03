package usecases

import (
	"context"
	"database/sql"
	"go-community/internal/common"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"strings"

	"github.com/google/uuid"
)

type EventRegistrationUsecase interface {
	Create(ctx context.Context, request models.CreateEventRegistrationRequest) (eventRegistration models.CreateEventRegistrationResponse, err error)
	GetRegistered(ctx context.Context, registeredBy string) (eventRegistrations []models.GetRegisteredResponse, err error)

	// Internal
	GetAll(ctx context.Context, params models.GetAllPaginationParams) (eventRegistrations []models.GetRegisteredResponse, err error)
	Verify(ctx context.Context, request models.VerifyRegistrationRequest, accountNumber string) (eventRegistration *models.EventRegistration, err error)
	Cancel(ctx context.Context, code string) (eventRegistration models.EventRegistration, err error)
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

func (eru *eventRegistrationUsecase) Create(ctx context.Context, request models.CreateEventRegistrationRequest, accountNumberOrigin string) (eventRegistration models.CreateEventRegistrationResponse, err error) {
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
	case strings.ToLower(session.Status) == "closed":
		err = models.ErrorRegistrationTimeDisabled
		return
	case strings.ToLower(session.Status) == "full":
		err = models.ErrorRegisterQuotaNotAvailable
		return
	case strings.ToLower(session.Status) == "walkin":
		err = models.ErrorNoRegistrationNeeded
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
		Name:                strings.ToUpper(request.Name),
		Identifier:          strings.ToLower(request.Identifier),
		Address:             request.Address,
		AccountNumber:       accountNumber,
		Code:                (uuid.New()).String(),
		EventCode:           event.Code,
		SessionCode:         session.Code,
		RegisteredBy:        strings.ToLower(request.Identifier),
		AccountNumberOrigin: accountNumberOrigin,
		Status:              "registered",
	}

	register = append(register, mainInput)

	for _, other := range request.Others {
		otherInput := models.EventRegistration{
			Name:                strings.ToUpper(other.Name),
			Address:             other.Address,
			Code:                (uuid.New()).String(),
			EventCode:           event.Code,
			SessionCode:         session.Code,
			RegisteredBy:        strings.ToLower(request.Identifier),
			AccountNumberOrigin: accountNumberOrigin,
			Status:              "registered",
		}

		register = append(register, otherInput)
	}

	if err = eru.rer.BulkCreate(ctx, &register); err != nil {
		return
	}

	// Update Session Quota & Session
	session.AvailableSeats -= countTotalRegister
	session.RegisteredSeats += countTotalRegister
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

func (eru *eventRegistrationUsecase) GetRegistered(ctx context.Context, registeredBy string, accountNumberOrigin string) (eventRegistrations []models.GetRegisteredResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	registers, err := eru.rer.GetSpecificByRegisteredBy(ctx, registeredBy, accountNumberOrigin)
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

func (eru *eventRegistrationUsecase) GetAll(ctx context.Context, params models.GetAllPaginationParams) (eventRegistrations []models.GetAllRegisteredResponse, count int64, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	registers, count, err := eru.rer.GetAllWithParams(ctx, params)
	if err != nil {
		return
	}

	response := make([]models.GetAllRegisteredResponse, len(registers))
	for i, p := range registers {
		response[i] = models.GetAllRegisteredResponse{
			Type:          models.TYPE_EVENT_REGISTRATION,
			Name:          p.EventRegistration.Name,
			Identifier:    p.EventRegistration.Identifier,
			Address:       p.EventRegistration.Address,
			AccountNumber: p.EventRegistration.AccountNumber,
			Code:          p.EventRegistration.Code,
			RegisteredBy:  p.EventRegistration.RegisteredBy,
			UpdatedBy:     p.EventRegistration.UpdatedBy,
			EventCode:     p.EventRegistration.EventCode,
			EventName:     p.EventGeneral.Name,
			SessionCode:   p.EventRegistration.SessionCode,
			SessionName:   p.EventSession.Name,
			Status:        p.EventRegistration.Status,
		}
	}

	return response, count, nil
}

func (eru *eventRegistrationUsecase) Verify(ctx context.Context, request models.VerifyRegistrationRequest, accountNumber string) (eventRegistration *models.EventRegistration, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	register, err := eru.rer.GetByCode(ctx, request.Code)
	if err != nil {
		return nil, err
	}

	if register.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	switch {
	case register.Status == "cancelled":
		return nil, models.ErrorRegistrationAlreadyCancel
	case register.Status == "verified":
		return nil, models.ErrorRegistrationAlreadyVerified
	case register.SessionCode != request.SessionCode:
		return nil, models.ErrorRegistrationWrongTime
	}

	// _, err := eru.egr.GetByCode(ctx, register.EventCode)
	// if err != nil {
	// 	return nil, err
	// }

	session, err := eru.esr.GetByCode(ctx, register.SessionCode)
	if err != nil {
		return nil, err
	}

	internalUser, err := eru.eur.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, err
	}

	register.Status = "verified"
	register.UpdatedBy = internalUser.PhoneNumber
	if internalUser.Email != "" {
		register.UpdatedBy = strings.ToLower(internalUser.Email)
	}

	session.ScannedSeats += 1

	if err := eru.rer.Update(ctx, register); err != nil {
		return nil, err
	}

	if err := eru.esr.Update(ctx, session); err != nil {
		return nil, err
	}

	return &register, nil
}

func (eru *eventRegistrationUsecase) Cancel(ctx context.Context, request models.CancelRegistrationRequest) (eventRegistration models.EventRegistration, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	register, err := eru.rer.GetByCode(ctx, request.Code)
	if err != nil {
		return
	}

	session, err := eru.esr.GetByCode(ctx, register.SessionCode)
	if err != nil {
		return
	}

	if register.DeletedAt.Valid {
		err = models.ErrorRegistrationAlreadyCancel
		return
	}

	if register.Status == "cancelled" {
		err = models.ErrorRegistrationAlreadyCancel
		return
	}

	switch {
	case register.DeletedAt.Valid:
		err = models.ErrorRegistrationAlreadyCancel
		return
	case register.Status == "verified":
		err = models.ErrorRegistrationAlreadyVerified
		return
	case register.Status == "cancelled":
		err = models.ErrorRegistrationAlreadyCancel
		return
	}

	register.Status = "cancelled"
	register.DeletedAt = sql.NullTime{Time: common.Now(), Valid: true}

	if err = eru.rer.Update(ctx, register); err != nil {
		return
	}

	// Update Session Quota & Session
	session.AvailableSeats += 1
	session.RegisteredSeats -= 1
	if session.AvailableSeats != 0 {
		session.Status = "active"
	}

	if err = eru.esr.Update(ctx, session); err != nil {
		return
	}

	return register, nil
}
