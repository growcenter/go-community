package pgsql

import "gorm.io/gorm"

type PostgreRepositories struct {
	Transaction           TransactionRepository
	Health                HealthRepository
	Campus                CampusRepository
	CoolCategory          CoolCategoryRepository
	Cool                  CoolRepository
	Location              LocationRepository
	User                  UserRepository
	UserRelation          UserRelationRepository
	EventUser             EventUserRepository
	EventGeneral          EventGeneralRepository
	EventSession          EventSessionRepository
	EventRegistration     EventRegistrationRepository
	EventCommunityRequest EventCommunityRequestRepository

	Role                    RoleRepository
	UserType                UserTypeRepository
	Event                   EventRepository
	EventInstance           EventInstanceRepository
	EventRegistrationRecord EventRegistrationRecordRepository
}

func New(db *gorm.DB) *PostgreRepositories {
	return &PostgreRepositories{
		Transaction:             NewTransactionRepository(db),
		Health:                  NewHealthRepository(db),
		Campus:                  NewCampusRepository(db, NewTransactionRepository(db)),
		CoolCategory:            NewCoolCategoryRepository(db, NewTransactionRepository(db)),
		Cool:                    NewCoolRepository(db, NewTransactionRepository(db)),
		Location:                NewLocationRepository(db, NewTransactionRepository(db)),
		User:                    NewUserRepository(db, NewTransactionRepository(db)),
		UserRelation:            NewUserRelationRepository(db, NewTransactionRepository(db)),
		EventUser:               NewEventUserRepository(db, NewTransactionRepository(db)),
		EventGeneral:            NewEventGeneralRepository(db, NewTransactionRepository(db)),
		EventSession:            NewEventSessionRepository(db, NewTransactionRepository(db)),
		EventRegistration:       NewEventRegistrationRepository(db, NewTransactionRepository(db)),
		EventCommunityRequest:   NewEventCommunityRequestRepository(db, NewTransactionRepository(db)),
		Role:                    NewRoleRepository(db, NewTransactionRepository(db)),
		UserType:                NewUserTypeRepository(db, NewTransactionRepository(db)),
		Event:                   NewEventRepository(db, NewTransactionRepository(db)),
		EventInstance:           NewEventInstanceRepository(db, NewTransactionRepository(db)),
		EventRegistrationRecord: NewEventRegistrationRecordRepository(db, NewTransactionRepository(db)),
	}
}
