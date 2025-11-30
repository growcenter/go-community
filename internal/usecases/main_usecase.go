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
	AccessRelationship      AccessRelationshipUsecase
}

func New(d Dependencies) *Usecases {
	healthUsecase := *NewHealthUsecase(d.Repository.Health)
	campusUsecase := *NewCampusUsecase(d.Repository.Campus)
	coolCategoryUsecase := *NewCoolCategoryUsecase(d.Repository.CoolCategory)
	locationUsecase := *NewLocationUsecase(d.Repository.Location, d.Repository.Campus)
	userUsecase := *NewUserUsecase(*d.Repository, *d.Config, *d.Authorization, d.Salt)
	eventCommunityRequestUsecase := *NewEventCommunityRequestUsecase(d.Repository.EventCommunityRequest, d.Repository.User)
	roleUsecase := *NewRoleUsecase(d.Repository.Role)
	userTypeUsecase := *NewUserTypeUsecase(*d.Repository)
	featureFlagUsecase := *NewFeatureFlagUsecase(*d.Repository)
	formAnswerUsecase := *NewFormAnswerUsecase(*d.Repository, *d.Config)
	formAssociationUsecase := NewFormAssociationUsecase(*d.Repository)
	eventAttendanceUsecase := *NewEventAttendanceUsecase(*d.Config, *d.Repository)
	formQuestionUsecase := *NewFormQuestionUsecase(*d.Repository)
	coolAttendanceUsecase := *NewCoolAttendanceUsecase(*d.Repository)
	configDBUsecase := *NewConfigDBUsecase(*d.Repository, *d.Config)
	eventRegistrationRecordUsecase := *NewEventRegistrationRecordUsecase(*d.Repository, *d.Config)
	accessRelationshipUsecase := NewAccessRelationshipUsecase(d.Repository.AccessRelationship)

	formUsecase := *NewFormUsecase(*d.Repository, &formQuestionUsecase)
	eventInstanceUsecase := *NewEventInstanceUsecase(*d.Config, *d.Authorization, *d.Repository, &formUsecase)
	eventUsecase := *NewEventUsecase(*d.Config, *d.Authorization, *d.Repository, &featureFlagUsecase, &eventInstanceUsecase, &formUsecase)
	eventStatusUsecase := *NewEventStatusUsecase(*d.Repository)
	eventRegistrationUsecase := *NewEventRegistrationUsecase(*d.Config, *d.Repository, &eventStatusUsecase, &formAnswerUsecase, &formAssociationUsecase, &formQuestionUsecase)
	coolUsecase := *NewCoolUsecase(*d.Repository, *d.Config, &featureFlagUsecase, *d.Indonesia)
	coolNewJoinerUsecase := *NewCoolNewJoinerUsecase(*d.Repository, d.Config, configDBUsecase)
	coolMeetingUsecase := *NewCoolMeetingUsecase(*d.Repository, *d.Config, &coolAttendanceUsecase)

	return &Usecases{
		Health:                  healthUsecase,
		Campus:                  campusUsecase,
		CoolCategory:            coolCategoryUsecase,
		Location:                locationUsecase,
		User:                    userUsecase,
		EventCommunityRequest:   eventCommunityRequestUsecase,
		Role:                    roleUsecase,
		UserType:                userTypeUsecase,
		Event:                   eventUsecase,
		EventRegistration:       eventRegistrationUsecase,
		EventAttendance:         eventAttendanceUsecase,
		EventRegistrationRecord: eventRegistrationRecordUsecase,
		EventInstance:           eventInstanceUsecase,
		FeatureFlag:             featureFlagUsecase,
		Config:                  configDBUsecase,
		Cool:                    coolUsecase,
		CoolNewJoiner:           coolNewJoinerUsecase,
		CoolMeeting:             coolMeetingUsecase,
		CoolAttendance:          coolAttendanceUsecase,
		Form:                    formUsecase,
		FormQuestion:            formQuestionUsecase,
		FormAnswer:              formAnswerUsecase,
		FormAssociation:         formAssociationUsecase,
		AccessRelationship:      accessRelationshipUsecase,
	}
}
