package usecases

import (
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/google"
	"go-community/internal/repositories/pgsql"
)

type Dependencies struct {
	Repository    *pgsql.PostgreRepositories
	Google        *google.GoogleAuth
	Authorization *authorization.Auth
	Salt          []byte
}

type Usecases struct {
	Health       healthUsecase
	Campus       campusUsecase
	CoolCategory coolCategoryUsecase
	Location     locationUsecase
	User         userUsecase
	EventUser    eventUserUsecase
	EventGeneral eventGeneralUsecase
}

func New(d Dependencies) *Usecases {
	health := NewHealthUsecase(d.Repository.Health)
	campus := NewCampusUsecase(d.Repository.Campus)
	coolCategory := NewCoolCategoryUsecase(d.Repository.CoolCategory)
	location := NewLocationUsecase(d.Repository.Location, d.Repository.Campus)
	user := NewUserUsecase(d.Repository.User, d.Repository.Campus, d.Repository.CoolCategory)
	eventUser := NewEventUserUsecase(d.Repository.EventUser, *d.Google, *d.Authorization, d.Salt)
	eventGeneral := NewEventGeneralUsecase(d.Repository.EventGeneral)

	return &Usecases{
		Health:       *health,
		Campus:       *campus,
		CoolCategory: *coolCategory,
		Location:     *location,
		User:         *user,
		EventUser:    *eventUser,
		EventGeneral: *eventGeneral,
	}
}
