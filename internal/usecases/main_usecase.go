package usecases

import (
	"go-community/internal/config"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/google"
	"go-community/internal/repositories/pgsql"
)

type Dependencies struct {
	Repository    *pgsql.PostgreRepositories
	Google        *google.GoogleAuth
	Authorization *authorization.Auth
	Salt          []byte
	Config        *config.Configuration
}

type Usecases struct {
	Health                  healthUsecase
	Campus                  campusUsecase
	CoolCategory            coolCategoryUsecase
	Location                locationUsecase
	User                    userUsecase
	EventUser               eventUserUsecase
	EventGeneral            eventGeneralUsecase
	EventSession            eventSessionUsecase
	EventRegistration       eventRegistrationUsecase
	EventCommunityRequest   eventCommunityRequestUsecase
	Role                    roleUsecase
	UserType                userTypeUsecase
	Event                   eventUsecase
	EventRegistrationRecord eventRegistrationRecordUsecase
	EventInstance           eventInstanceUsecase
}

func New(d Dependencies) *Usecases {
	return &Usecases{
		Health:                  *NewHealthUsecase(d.Repository.Health),
		Campus:                  *NewCampusUsecase(d.Repository.Campus),
		CoolCategory:            *NewCoolCategoryUsecase(d.Repository.CoolCategory),
		Location:                *NewLocationUsecase(d.Repository.Location, d.Repository.Campus),
		User:                    *NewUserUsecase(d.Repository.User, d.Repository.UserRelation, d.Repository.Campus, d.Repository.CoolCategory, d.Repository.Cool, d.Repository.UserType, d.Repository.Role, *d.Repository, *d.Config, *d.Authorization, d.Salt),
		EventUser:               *NewEventUserUsecase(d.Repository.EventUser, *d.Google, *d.Authorization, d.Salt),
		EventGeneral:            *NewEventGeneralUsecase(d.Repository.EventGeneral),
		EventSession:            *NewEventSessionUsecase(d.Repository.EventSession, d.Repository.EventGeneral),
		EventRegistration:       *NewEventRegistrationUsecase(d.Repository.EventRegistration, d.Repository.EventGeneral, d.Repository.EventSession, d.Repository.EventUser),
		EventCommunityRequest:   *NewEventCommunityRequestUsecase(d.Repository.EventCommunityRequest, d.Repository.EventUser),
		Role:                    *NewRoleUsecase(d.Repository.Role),
		UserType:                *NewUserTypeUsecase(*d.Repository),
		Event:                   *NewEventUsecase(*d.Config, *d.Authorization, *d.Repository),
		EventRegistrationRecord: *NewEventRegistrationRecordUsecase(*d.Repository),
		EventInstance:           *NewEventInstanceUsecase(*d.Config, *d.Authorization, *d.Repository),
	}
}
