package pgsql

import "gorm.io/gorm"

type PostgreRepositories struct {
	Transaction  TransactionRepository
	Health       HealthRepository
	Campus       CampusRepository
	CoolCategory CoolCategoryRepository
	Location     LocationRepository
	User         UserRepository
	EventUser    EventUserRepository
}

func New(db *gorm.DB) *PostgreRepositories {
	return &PostgreRepositories{
		Transaction:  NewTransactionRepository(db),
		Health:       NewHealthRepository(db),
		Campus:       NewCampusRepository(db, NewTransactionRepository(db)),
		CoolCategory: NewCoolCategoryRepository(db, NewTransactionRepository(db)),
		Location:     NewLocationRepository(db, NewTransactionRepository(db)),
		User:         NewUserRepository(db, NewTransactionRepository(db)),
		EventUser:    NewEventUserRepository(db, NewTransactionRepository(db)),
	}
}
