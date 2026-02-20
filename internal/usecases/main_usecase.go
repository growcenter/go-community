package usecases

import (
	"go-community/internal/config"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/cache"
	"go-community/internal/pkg/google"
	"go-community/internal/repositories/pgsql"
)

type Dependencies struct {
	Repository    *pgsql.PostgreRepositories
	Google        *google.GoogleAuth
	Authorization *authorization.Auth
	Salt          []byte
	Config        *config.Configuration
	Cache         *cache.MemoryCache

	Event        EventUsecase
	EventSession EventSessionUsecase

	Form            FormUsecase
	FormQuestion    FormQuestionUsecase
	FormAnswer      FormAnswerUsecase
	FormAssociation FormAssociationUsecase
}

type Usecases struct {
	Health                healthUsecase
	Campus                campusUsecase
	CoolCategory          coolCategoryUsecase
	Location              locationUsecase
	User                  userUsecase
	EventCommunityRequest eventCommunityRequestUsecase
	Role                  roleUsecase
	UserType              userTypeUsecase

	Event         EventUsecase
	EventSession  EventSessionUsecase
	FeatureFlag   featureFlagUsecase
	Config        configDBUsecase
	Cool          coolUsecase
	CoolNewJoiner coolNewJoinerUsecase

	Form            FormUsecase
	FormQuestion    FormQuestionUsecase
	FormAnswer      FormAnswerUsecase
	FormAssociation FormAssociationUsecase
}

func New(d Dependencies) *Usecases {
	// Initialise leaf usecases first (no cross-usecase dependencies)
	eventUsecase := NewEventUsecase(&d)
	eventSessionUsecase := NewEventSessionUsecase(&d)
	formUsecase := NewFormUsecase(&d)
	formQuestionUsecase := NewFormQuestionUsecase(&d)
	formAnswerUsecase := NewFormAnswerUsecase(&d)
	formAssociationUsecase := NewFormAssociationUsecase(&d)

	// Inject into dependencies so cross-usecase calls (e.g., event → session) work
	d.EventSession = eventSessionUsecase
	d.Event = eventUsecase
	d.Form = formUsecase
	d.FormQuestion = formQuestionUsecase
	d.FormAnswer = formAnswerUsecase
	d.FormAssociation = formAssociationUsecase

	return &Usecases{
		Health:                *NewHealthUsecase(d.Repository.Health),
		Campus:                *NewCampusUsecase(d.Repository.Campus),
		CoolCategory:          *NewCoolCategoryUsecase(d.Repository.CoolCategory),
		Location:              *NewLocationUsecase(d.Repository.Location, d.Repository.Campus),
		User:                  *NewUserUsecase(d.Repository.User, d.Repository.UserRelation, d.Repository.Campus, d.Repository.CoolCategory, d.Repository.Cool, d.Repository.UserType, d.Repository.Role, *d.Repository, *d.Config, *d.Authorization, d.Salt),
		EventCommunityRequest: *NewEventCommunityRequestUsecase(d.Repository.EventCommunityRequest, d.Repository.User),
		Role:                  *NewRoleUsecase(d.Repository.Role),
		UserType:              *NewUserTypeUsecase(*d.Repository),
		Event:                 eventUsecase,
		EventSession:          eventSessionUsecase,
		FeatureFlag:           *NewFeatureFlagUsecase(*d.Repository),
		Config:                *NewConfigDBUsecase(*d.Repository, *d.Config),
		Cool:                  *NewCoolUsecase(*d.Repository, *d.Config, &featureFlagUsecase{r: *d.Repository}),
		CoolNewJoiner:         *NewCoolNewJoinerUsecase(*d.Repository, d.Config, configDBUsecase{r: *d.Repository}),
		Form:                  formUsecase,
		FormQuestion:          formQuestionUsecase,
		FormAnswer:            formAnswerUsecase,
		FormAssociation:       formAssociationUsecase,
	}
}
