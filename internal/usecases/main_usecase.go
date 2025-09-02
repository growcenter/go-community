package usecases

import (
	indonesiaAPI "go-community/internal/clients/indonesia-api"
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
	Indonesia     *indonesiaAPI.Client
}

type Usecases struct {
	Health                  healthUsecase
	Campus                  campusUsecase
	CoolCategory            coolCategoryUsecase
	Location                locationUsecase
	User                    userUsecase
	EventCommunityRequest   eventCommunityRequestUsecase
	Role                    roleUsecase
	UserType                userTypeUsecase
	Event                   eventUsecase
	EventRegistrationRecord eventRegistrationRecordUsecase
	EventInstance           eventInstanceUsecase
	FeatureFlag             featureFlagUsecase
	Config                  configDBUsecase
	Cool                    coolUsecase
	CoolNewJoiner           coolNewJoinerUsecase
	CoolMeeting             coolMeetingUsecase
	CoolAttendance          coolAttendanceUsecase
}

func New(d Dependencies) *Usecases {
	return &Usecases{
		Health:                  *NewHealthUsecase(d.Repository.Health),
		Campus:                  *NewCampusUsecase(d.Repository.Campus),
		CoolCategory:            *NewCoolCategoryUsecase(d.Repository.CoolCategory),
		Location:                *NewLocationUsecase(d.Repository.Location, d.Repository.Campus),
		User:                    *NewUserUsecase(d.Repository.User, d.Repository.UserRelation, d.Repository.Campus, d.Repository.CoolCategory, d.Repository.Cool, d.Repository.UserType, d.Repository.Role, *d.Repository, *d.Config, *d.Authorization, d.Salt),
		EventCommunityRequest:   *NewEventCommunityRequestUsecase(d.Repository.EventCommunityRequest, d.Repository.User),
		Role:                    *NewRoleUsecase(d.Repository.Role),
		UserType:                *NewUserTypeUsecase(*d.Repository),
		Event:                   *NewEventUsecase(*d.Config, *d.Authorization, *d.Repository, &featureFlagUsecase{r: *d.Repository}),
		EventRegistrationRecord: *NewEventRegistrationRecordUsecase(*d.Repository, *d.Config),
		EventInstance:           *NewEventInstanceUsecase(*d.Config, *d.Authorization, *d.Repository),
		FeatureFlag:             *NewFeatureFlagUsecase(*d.Repository),
		Config:                  *NewConfigDBUsecase(*d.Repository, *d.Config),
		Cool:                    *NewCoolUsecase(*d.Repository, *d.Config, &featureFlagUsecase{r: *d.Repository}, *d.Indonesia),
		CoolNewJoiner:           *NewCoolNewJoinerUsecase(*d.Repository, d.Config, configDBUsecase{r: *d.Repository}),
		CoolMeeting:             *NewCoolMeetingUsecase(*d.Repository, *d.Config, &coolAttendanceUsecase{r: *d.Repository}),
		CoolAttendance:          *NewCoolAttendanceUsecase(*d.Repository),
	}
}
