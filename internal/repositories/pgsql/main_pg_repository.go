package pgsql

import "gorm.io/gorm"

type PostgreRepositories struct {
	Transaction		TransactionRepository
	Health 			HealthRepository
	Campus 			CampusRepository
	CoolCategory	CoolCategoryRepository
}

func New(db *gorm.DB) *PostgreRepositories {
	return &PostgreRepositories{
		Transaction: NewTransactionRepository(db),
		Health: NewHealthRepository(db),
		Campus: NewCampusRepository(db, NewTransactionRepository(db)),
		CoolCategory: NewCoolCategoryRepository(db, NewTransactionRepository(db)),
	}
}