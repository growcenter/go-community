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

	Event         EventUsecase
	EventInstance EventInstanceUsecase
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

	Event                   EventUsecase
	EventInstance           EventInstanceUsecase
	EventRegistrationRecord eventRegistrationRecordUsecase
	FeatureFlag             featureFlagUsecase
	Config                  configDBUsecase
	Cool                    coolUsecase
	CoolNewJoiner           coolNewJoinerUsecase
}

func New(d Dependencies) *Usecases {
	// Initialize EventInstance usecase first
	eventUsecase := NewEventUsecase(&d)
	eventInstanceUsecase := NewEventInstanceUsecase(&d)

	// Inject into dependencies
	d.EventInstance = eventInstanceUsecase
	d.Event = eventUsecase

	return &Usecases{
		Health:                  *NewHealthUsecase(d.Repository.Health),
		Campus:                  *NewCampusUsecase(d.Repository.Campus),
		CoolCategory:            *NewCoolCategoryUsecase(d.Repository.CoolCategory),
		Location:                *NewLocationUsecase(d.Repository.Location, d.Repository.Campus),
		User:                    *NewUserUsecase(d.Repository.User, d.Repository.UserRelation, d.Repository.Campus, d.Repository.CoolCategory, d.Repository.Cool, d.Repository.UserType, d.Repository.Role, *d.Repository, *d.Config, *d.Authorization, d.Salt),
		EventCommunityRequest:   *NewEventCommunityRequestUsecase(d.Repository.EventCommunityRequest, d.Repository.User),
		Role:                    *NewRoleUsecase(d.Repository.Role),
		UserType:                *NewUserTypeUsecase(*d.Repository),
		Event:                   eventUsecase,
		EventInstance:           eventInstanceUsecase,
		EventRegistrationRecord: *NewEventRegistrationRecordUsecase(*d.Repository, *d.Config),
		FeatureFlag:             *NewFeatureFlagUsecase(*d.Repository),
		Config:                  *NewConfigDBUsecase(*d.Repository, *d.Config),
		Cool:                    *NewCoolUsecase(*d.Repository, *d.Config, &featureFlagUsecase{r: *d.Repository}),
		CoolNewJoiner:           *NewCoolNewJoinerUsecase(*d.Repository, d.Config, configDBUsecase{r: *d.Repository}),
	}
}
