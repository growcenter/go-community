package pgsql

import "gorm.io/gorm"

type PostgreRepositories struct {
	Transaction               TransactionRepository
	Health                    HealthRepository
	Campus                    CampusRepository
	CoolCategory              CoolCategoryRepository
	Cool                      CoolRepository
	Location                  LocationRepository
	User                      UserRepository
	UserRelation              UserRelationRepository
	EventCommunityRequest     EventCommunityRequestRepository
	EventRegistration         EventRegistrationRepository
	EventAttendance           EventAttendanceRepository

	FeatureFlag FeatureFlagRepository
	Config      ConfigRepository

	Role                    RoleRepository
	UserType                UserTypeRepository
	Event                   EventRepository
	EventInstance           EventInstanceRepository
	EventRegistrationRecord EventRegistrationRecordRepository
	EventQuestion           EventQuestionRepository
	CoolNewJoiner           CoolNewJoinerRepository

	CoolMeeting    CoolMeetingRepository
	CoolAttendance CoolAttendanceRepository
	Form           FormRepository
	FormQuestion   FormQuestionRepository
	FormAnswer     FormAnswerRepository
	FormAssociation FormAssociationRepository
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
		EventCommunityRequest:   NewEventCommunityRequestRepository(db, NewTransactionRepository(db)),
		EventRegistration:       NewEventRegistrationRepository(db),
		EventAttendance:         NewEventAttendanceRepository(db),
		Role:                    NewRoleRepository(db, NewTransactionRepository(db)),
		UserType:                NewUserTypeRepository(db, NewTransactionRepository(db)),
		Event:                   NewEventRepository(db, NewTransactionRepository(db)),
		EventInstance:           NewEventInstanceRepository(db, NewTransactionRepository(db)),
		EventRegistrationRecord: NewEventRegistrationRecordRepository(db, NewTransactionRepository(db)),
		EventQuestion:           NewEventQuestionRepository(db),
		FeatureFlag:             NewFeatureFlagRepository(db),
		CoolNewJoiner:           NewCoolNewJoinerRepository(db),
		Config:                  NewConfigRepository(db),
		CoolMeeting:             NewCoolMeetingRepository(db),
		CoolAttendance:          NewCoolAttendanceRepository(db),
		Form:                    NewFormRepository(db),
		FormQuestion:            NewFormQuestionRepository(db),
		FormAnswer:              NewFormAnswerRepository(db),
		FormAssociation:         NewFormAssociationRepository(db),
	}
}
