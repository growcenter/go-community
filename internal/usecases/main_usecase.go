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
	Health            healthUsecase
	Campus            campusUsecase
	CoolCategory      coolCategoryUsecase
	Location          locationUsecase
	User              userUsecase
	EventUser         eventUserUsecase
	EventGeneral      eventGeneralUsecase
	EventSession      eventSessionUsecase
	EventRegistration eventRegistrationUsecase
}

func New(d Dependencies) *Usecases {
	return &Usecases{
		Health:            *NewHealthUsecase(d.Repository.Health),
		Campus:            *NewCampusUsecase(d.Repository.Campus),
		CoolCategory:      *NewCoolCategoryUsecase(d.Repository.CoolCategory),
		Location:          *NewLocationUsecase(d.Repository.Location, d.Repository.Campus),
		User:              *NewUserUsecase(d.Repository.User, d.Repository.Campus, d.Repository.CoolCategory),
		EventUser:         *NewEventUserUsecase(d.Repository.EventUser, *d.Google, *d.Authorization, d.Salt),
		EventGeneral:      *NewEventGeneralUsecase(d.Repository.EventGeneral),
		EventSession:      *NewEventSessionUsecase(d.Repository.EventSession, d.Repository.EventGeneral),
		EventRegistration: *NewEventRegistrationUsecase(d.Repository.EventRegistration, d.Repository.EventGeneral, d.Repository.EventSession, d.Repository.EventUser),
	}
}
