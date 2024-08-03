package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/generator"
	"go-community/internal/pkg/google"
	"go-community/internal/pkg/hash"
	"go-community/internal/repositories/pgsql"
	"net/http"
	"strings"
)

type EventUserUsecase interface {
	Redirect(ctx context.Context) (url string, err error)
	Account(ctx context.Context, state string, code string) (eventUser *models.EventUser, token string, statusCode int, err error)
	ManualRegister(ctx context.Context, request models.CreateEventUserManualRequest) (eventUser *models.EventUser, token string, err error)
	ManualLogin(ctx context.Context, request models.LoginEventUserManualRequest) (eventUser *models.EventUser, token string, err error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (eventUser *models.EventUser, err error)
	UpdateRole(ctx context.Context, request models.UpdateAccountRoleRequest) (response *models.UpdateAccountRoleResponse, err error)
	Logout(ctx context.Context, accountNumber string) (token string, isLoggedIn bool, err error)
	UpdatePassword(ctx context.Context, request models.UpdatePasswordRequest) (eventUser models.EventUser, err error)
}

type eventUserUsecase struct {
	eur pgsql.EventUserRepository
	g   google.GoogleAuth
	a   authorization.Auth
	s   []byte
}

func NewEventUserUsecase(eur pgsql.EventUserRepository, g google.GoogleAuth, a authorization.Auth, s []byte) *eventUserUsecase {
	return &eventUserUsecase{
		eur: eur,
		g:   g,
		a:   a,
		s:   s,
	}
}

func (euu *eventUserUsecase) Redirect(ctx context.Context) (authUrl string, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	url := euu.g.Redirect()

	return url, nil
}

func (euu *eventUserUsecase) Account(ctx context.Context, state string, code string) (eventUser *models.EventUser, token string, statusCode int, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	googleData, err := euu.g.Fetch(state, code)
	if err != nil {
		return nil, "", http.StatusInternalServerError, models.ErrorFetchGoogle
	}

	exist, err := euu.eur.GetByEmail(ctx, googleData.Email)
	if err != nil {
		return nil, "", http.StatusInternalServerError, err
	}

	if exist.ID != 0 {
		exist.Role = "user"
		tokenStatus := "active"
		bearerToken, err := euu.a.Generate(exist.AccountNumber, exist.Role, tokenStatus)
		if err != nil {
			return nil, "", http.StatusInternalServerError, err
		}

		return &exist, bearerToken, http.StatusOK, nil
	}

	accountNumber, err := generator.AccountNumber()
	if err != nil {
		return nil, "", http.StatusInternalServerError, err
	}

	input := models.EventUser{
		Name:          googleData.Name,
		Email:         strings.ToLower(googleData.Email),
		Role:          "user",
		State:         "1",
		Status:        "active",
		AccountNumber: accountNumber,
	}

	if err := euu.eur.Create(ctx, &input); err != nil {
		return nil, "", http.StatusInternalServerError, err
	}

	tokenStatus := "active"
	bearerToken, err := euu.a.Generate(accountNumber, input.Role, tokenStatus)
	if err != nil {
		return nil, "", http.StatusInternalServerError, err
	}

	return &input, bearerToken, http.StatusCreated, nil
}

func (euu *eventUserUsecase) ManualRegister(ctx context.Context, request models.CreateEventUserManualRequest) (eventUser *models.EventUser, token string, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	switch {
	case request.Email != "" && request.PhoneNumber != "":
		existEmail, err := euu.eur.GetByEmail(ctx, strings.ToLower(request.Email))
		if err != nil {
			return nil, "", err
		}

		existPhone, err := euu.eur.GetByPhoneNumber(ctx, strings.ToLower(request.PhoneNumber))
		if err != nil {
			return nil, "", err
		}

		if existEmail.ID != 0 || existPhone.ID != 0 || strings.ToLower(request.Email) == strings.ToLower(existEmail.Email) || request.PhoneNumber == existPhone.PhoneNumber {
			return nil, "", models.ErrorAlreadyExist
		}

		accountNumber, err := generator.AccountNumber()
		if err != nil {
			return nil, "", err
		}

		salted := append([]byte(request.Password), euu.s...)
		password, err := hash.Generate(salted)
		if err != nil {
			return nil, "", err
		}

		input := models.EventUser{
			Name:          request.Name,
			Email:         strings.ToLower(request.Email),
			PhoneNumber:   request.PhoneNumber,
			Role:          "user",
			State:         "1",
			Status:        "active",
			AccountNumber: accountNumber,
			Password:      password,
		}

		if err := euu.eur.Create(ctx, &input); err != nil {
			return nil, "", err
		}

		tokenStatus := "active"
		bearerToken, err := euu.a.Generate(accountNumber, input.Role, tokenStatus)
		if err != nil {
			return nil, "", err
		}

		return &input, bearerToken, nil
	case request.Email != "" && request.PhoneNumber == "":
		exist, err := euu.eur.GetByEmail(ctx, strings.ToLower(request.Email))
		if err != nil {
			return nil, "", err
		}

		if exist.ID != 0 || strings.ToLower(request.Email) == strings.ToLower(exist.Email) {
			return nil, "", models.ErrorAlreadyExist
		}

		accountNumber, err := generator.AccountNumber()
		if err != nil {
			return nil, "", err
		}

		salted := append([]byte(request.Password), euu.s...)
		password, err := hash.Generate(salted)
		if err != nil {
			return nil, "", err
		}

		input := models.EventUser{
			Name:          request.Name,
			Email:         strings.ToLower(request.Email),
			Role:          "user",
			State:         "1",
			Status:        "active",
			AccountNumber: accountNumber,
			Password:      password,
		}

		if err := euu.eur.Create(ctx, &input); err != nil {
			return nil, "", err
		}

		tokenStatus := "active"
		bearerToken, err := euu.a.Generate(accountNumber, input.Role, tokenStatus)
		if err != nil {
			return nil, "", err
		}

		return &input, bearerToken, nil
	case request.Email == "" && request.PhoneNumber != "":
		exist, err := euu.eur.GetByPhoneNumber(ctx, strings.ToLower(request.PhoneNumber))
		if err != nil {
			return nil, "", err
		}

		if exist.ID != 0 || request.PhoneNumber == exist.PhoneNumber {
			return nil, "", models.ErrorAlreadyExist
		}

		accountNumber, err := generator.AccountNumber()
		if err != nil {
			return nil, "", err
		}

		salted := append([]byte(request.Password), euu.s...)
		password, err := hash.Generate(salted)
		if err != nil {
			return nil, "", err
		}

		input := models.EventUser{
			Name:          request.Name,
			PhoneNumber:   request.PhoneNumber,
			Role:          "user",
			State:         "1",
			Status:        "active",
			AccountNumber: accountNumber,
			Password:      password,
		}

		if err := euu.eur.Create(ctx, &input); err != nil {
			return nil, "", err
		}

		tokenStatus := "active"
		bearerToken, err := euu.a.Generate(accountNumber, input.Role, tokenStatus)
		if err != nil {
			return nil, "", err
		}

		return &input, bearerToken, nil
	default:
		return nil, "", models.ErrorEmailPhoneNumberEmpty
	}
}

func (euu *eventUserUsecase) ManualLogin(ctx context.Context, request models.LoginEventUserManualRequest) (eventUser *models.EventUser, token string, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	user, err := euu.eur.GetByEmailPhone(ctx, strings.ToLower(request.Identifier))
	if err != nil {
		return nil, "", err
	}

	if user.ID == 0 {
		return nil, "", models.ErrorUserNotFound
	}

	salted := append([]byte(request.Password), euu.s...)
	if err = hash.Validate(user.Password, string(salted)); err != nil {
		return nil, "", models.ErrorInvalidPassword
	}

	tokenStatus := "active"
	bearerToken, err := euu.a.Generate(user.AccountNumber, strings.ToLower(user.Role), tokenStatus)
	if err != nil {
		return nil, "", err
	}

	return &user, bearerToken, nil
}

func (euu *eventUserUsecase) GetByAccountNumber(ctx context.Context, accountNumber string) (eventUser *models.EventUser, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	user, err := euu.eur.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, err
	}

	if user.ID == 0 {
		return nil, models.ErrorUserNotFound
	}

	return &user, nil
}

func (euu *eventUserUsecase) UpdateRole(ctx context.Context, request models.UpdateAccountRoleRequest) (response *models.UpdateAccountRoleResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	for _, accountNumber := range request.AccountNumbers {
		user, err := euu.eur.GetByAccountNumber(ctx, accountNumber)
		if err != nil {
			return nil, err
		}

		if user.ID == 0 {
			return nil, models.ErrorUserNotFound
		}
	}

	if err := euu.eur.BulkUpateRoleByAccountNumbers(ctx, request.AccountNumbers, strings.ToLower(request.Role)); err != nil {
		return nil, err
	}

	response = &models.UpdateAccountRoleResponse{
		Type:           models.TYPE_EVENT_REGISTRATION,
		AccountNumbers: request.AccountNumbers,
		Role:           strings.ToLower(request.Role),
	}

	return response, nil
}

func (euu *eventUserUsecase) Logout(ctx context.Context, accountNumber string) (response *models.LogoutEventUserResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	user, err := euu.eur.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		return
	}

	if user.ID == 0 {
		err = models.ErrorUserNotFound
		return
	}

	tokenStatus := "inactive"
	token, err := euu.a.Generate(user.AccountNumber, strings.ToLower(user.Role), tokenStatus)
	if err != nil {
		return
	}

	response = &models.LogoutEventUserResponse{
		Type:          models.TYPE_EVENT_USER,
		AccountNumber: accountNumber,
		Token:         token,
		IsLoggedOut:   true,
	}

	return response, nil
}

func (euu *eventUserUsecase) UpdatePassword(ctx context.Context, request models.UpdatePasswordRequest) (eventUser models.EventUser, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	user, err := euu.eur.GetByEmailPhone(ctx, strings.ToLower(request.Identifier))
	if err != nil {
		return
	}

	if user.ID == 0 {
		err = models.ErrorUserNotFound
		return
	}

	salted := append([]byte(request.Password), euu.s...)
	password, err := hash.Generate(salted)
	if err != nil {
		return
	}

	user.Password = password
	if err = euu.eur.Update(ctx, &user); err != nil {
		return
	}

	return user, nil
}
