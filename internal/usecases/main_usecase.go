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
	EventRegistration       eventRegistrationUsecase
	EventAttendance         eventAttendanceUsecase
	EventRegistrationRecord eventRegistrationRecordUsecase
	EventInstance           eventInstanceUsecase
	FeatureFlag             featureFlagUsecase
	Config                  configDBUsecase
	Cool                    coolUsecase
	CoolNewJoiner           coolNewJoinerUsecase
	CoolMeeting             coolMeetingUsecase
	CoolAttendance          coolAttendanceUsecase
	Form                    formUsecase
	FormQuestion            formQuestionUsecase
	FormAnswer              formAnswerUsecase
	FormAssociation         formAssociationUsecase
}

func New(d Dependencies) *Usecases {
	formQuestionUsecase := *NewFormQuestionUsecase(*d.Repository)
	formUsecase := *NewFormUsecase(*d.Repository, &formQuestionUsecase)
	formAnswerUsecase := *NewFormAnswerUsecase(*d.Repository, *d.Config)
	formAssociationUsecase := NewFormAssociationUsecase(*d.Repository)
	featureFlagUsecase := *NewFeatureFlagUsecase(*d.Repository)
	eventInstanceUsecase := *NewEventInstanceUsecase(*d.Config, *d.Authorization, *d.Repository)
	coolAttendanceUsecase := *NewCoolAttendanceUsecase(*d.Repository)
	configDBUsecase := *NewConfigDBUsecase(*d.Repository, *d.Config)

	return &Usecases{
		Health:                  *NewHealthUsecase(d.Repository.Health),
		Campus:                  *NewCampusUsecase(d.Repository.Campus),
		CoolCategory:            *NewCoolCategoryUsecase(d.Repository.CoolCategory),
		Location:                *NewLocationUsecase(d.Repository.Location, d.Repository.Campus),
		User:                    *NewUserUsecase(*d.Repository, *d.Config, *d.Authorization, d.Salt),
		EventCommunityRequest:   *NewEventCommunityRequestUsecase(d.Repository.EventCommunityRequest, d.Repository.User),
		Role:                    *NewRoleUsecase(d.Repository.Role),
		UserType:                *NewUserTypeUsecase(*d.Repository),
		Event:                   *NewEventUsecase(*d.Config, *d.Authorization, *d.Repository, &featureFlagUsecase, &eventInstanceUsecase, &formUsecase),
		EventRegistration:       *NewEventRegistrationUsecase(*d.Config, *d.Repository, &formAnswerUsecase, &formAssociationUsecase),
		EventAttendance:         *NewEventAttendanceUsecase(*d.Config, *d.Repository),
		EventRegistrationRecord: *NewEventRegistrationRecordUsecase(*d.Repository, *d.Config),
		EventInstance:           eventInstanceUsecase,
		FeatureFlag:             featureFlagUsecase,
		Config:                  configDBUsecase,
		Cool:                    *NewCoolUsecase(*d.Repository, *d.Config, &featureFlagUsecase, *d.Indonesia),
		CoolNewJoiner:           *NewCoolNewJoinerUsecase(*d.Repository, d.Config, configDBUsecase),
		CoolMeeting:             *NewCoolMeetingUsecase(*d.Repository, *d.Config, &coolAttendanceUsecase),
		CoolAttendance:          coolAttendanceUsecase,
		Form:                    formUsecase,
		FormQuestion:            formQuestionUsecase,
		FormAnswer:              formAnswerUsecase,
		FormAssociation:         formAssociationUsecase,
	}
}
