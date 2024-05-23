package pgsql

import (
	"fmt"
	"go-community/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	Transaction(fc func(dtx *gorm.DB) error) error
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

// ExecuteInTransaction is a utility function to handle GORM transactions.
// Reference: https://gorm.io/docs/transactions.html#A-Specific-Example
func (tr *transactionRepository) Transaction(fc func(dtx *gorm.DB) error) error {
	dtx := tr.db.Begin()
	if dtx.Error != nil {
		return dtx.Error
	}

	logger.Logger.Info("[DATABASE-TRX-BEGIN]")

	defer func() {
		if r := recover(); r != nil {
			dtx.Rollback()
			err := fmt.Errorf("[DATABASE-ERROR] panic happened because: " + fmt.Sprintf("%v", r))
			logger.Logger.Error("[DATABASE-TRX-PANIC]", zap.Error(err))
		}
	}()

	if err := fc(dtx); err != nil {
		dtx.Rollback()
		logger.Logger.Error("[DATABASE-TRX-ROLLBACK]", zap.Error(err))
		return err
	}

	return dtx.Commit().Error
}