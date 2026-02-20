package pgsql

import (
	"context"

	"gorm.io/gorm"
)

// BaseRepository contains common dependencies for all repositories
// This eliminates the need to declare `db *gorm.DB` in every repository
type BaseRepository struct {
	DB *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{DB: db}
}

// db returns the appropriate *gorm.DB for the current context.
// If an Atomic transaction is active (stored in ctx by Atomic), it returns that
// transactional DB so all writes join the enclosing transaction automatically.
// Otherwise it falls back to the base DB connection for standalone operations.
func (b *BaseRepository) db(ctx context.Context) *gorm.DB {
	if tx, ok := TxFromContext(ctx); ok {
		return tx.WithContext(ctx)
	}
	return b.DB.WithContext(ctx)
}

type PostgreRepositories struct {
	Transaction           TransactionRepository
	Health                HealthRepository
	Campus                CampusRepository
	CoolCategory          CoolCategoryRepository
	Cool                  CoolRepository
	Location              LocationRepository
	User                  UserRepository
	UserRelation          UserRelationRepository
	EventCommunityRequest EventCommunityRequestRepository

	FeatureFlag FeatureFlagRepository
	Config      ConfigRepository

	Role                    RoleRepository
	UserType                UserTypeRepository
	Event                   EventRepository
	EventSession            EventSessionRepository
	EventInstance           EventInstanceRepository
	EventRegistrationRecord EventRegistrationRecordRepository
	CoolNewJoiner           CoolNewJoinerRepository

	Form            FormRepository
	FormQuestion    FormQuestionRepository
	FormAnswer      FormAnswerRepository
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
		Role:                    NewRoleRepository(db, NewTransactionRepository(db)),
		UserType:                NewUserTypeRepository(db, NewTransactionRepository(db)),
		Event:                   NewEventRepository(db),
		EventInstance:           NewEventInstanceRepository(db),
		EventRegistrationRecord: NewEventRegistrationRecordRepository(db, NewTransactionRepository(db)),
		FeatureFlag:             NewFeatureFlagRepository(db),
		CoolNewJoiner:           NewCoolNewJoinerRepository(db),
		Config:                  NewConfigRepository(db),
		Form:                    NewFormRepository(db),
		FormQuestion:            NewFormQuestionRepository(db),
		FormAnswer:              NewFormAnswerRepository(db),
		FormAssociation:         NewFormAssociationRepository(db),
	}
}
