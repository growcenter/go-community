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
	googleData, err := euu.g.Fetch(state, code)
	if err != nil {
		return nil, "", http.StatusInternalServerError, models.ErrorFetchGoogle
	}

	exist, err := euu.eur.GetByEmail(ctx, googleData.Email)
	if err != nil {
		return nil, "", http.StatusInternalServerError, err
	}

	if exist.ID != 0 {
		bearerToken, err := euu.a.Generate(exist.AccountNumber)
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
		Role:          "USER",
		State:         "1",
		Status:        "active",
		AccountNumber: accountNumber,
	}

	if err := euu.eur.Create(ctx, &input); err != nil {
		return nil, "", http.StatusInternalServerError, err
	}

	bearerToken, err := euu.a.Generate(accountNumber)
	if err != nil {
		return nil, "", http.StatusInternalServerError, err
	}

	return &input, bearerToken, http.StatusCreated, nil
}

func (euu *eventUserUsecase) ManualRegister(ctx context.Context, request models.CreateEventUserManualRequest) (eventUser *models.EventUser, token string, err error) {
	switch {
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
			Role:          "USER",
			State:         "1",
			Status:        "active",
			AccountNumber: accountNumber,
			Password:      password,
		}

		if err := euu.eur.Create(ctx, &input); err != nil {
			return nil, "", err
		}

		bearerToken, err := euu.a.Generate(accountNumber)
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
			Role:          "USER",
			State:         "1",
			Status:        "active",
			AccountNumber: accountNumber,
			Password:      password,
		}

		if err := euu.eur.Create(ctx, &input); err != nil {
			return nil, "", err
		}

		bearerToken, err := euu.a.Generate(accountNumber)
		if err != nil {
			return nil, "", err
		}

		return &input, bearerToken, nil
	default:
		return nil, "", models.ErrorEmailPhoneNumberEmpty
	}
}

func (euu *eventUserUsecase) ManualLogin(ctx context.Context, request models.LoginEventUserManualRequest) (eventUser *models.EventUser, token string, err error) {
	user, err := euu.eur.GetByEmailPhone(ctx, request.Identifier)
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

	bearerToken, err := euu.a.Generate(user.AccountNumber)
	if err != nil {
		return nil, "", err
	}

	return &user, bearerToken, nil
}
