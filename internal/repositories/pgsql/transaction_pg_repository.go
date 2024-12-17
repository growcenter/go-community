package pgsql

import (
	"context"
	"fmt"
	"go-community/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	Transaction(fc func(dtx *gorm.DB) error) error
	Atomic(ctx context.Context, fc func(ctx context.Context, r *PostgreRepositories) error) error
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

func (tr *transactionRepository) Atomic(ctx context.Context, fc func(ctx context.Context, r *PostgreRepositories) error) error {
	tx := tr.db.WithContext(ctx).Begin()
	err := tx.Error
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Create a repository bound to the transaction
	// Ensure proper rollback or commit
	defer func() {
		if r := recover(); r != nil {
			// Handle panic and rollback
			tx.Rollback()
			err := fmt.Errorf("[DATABASE-ERROR] panic happened because: " + fmt.Sprintf("%v", r))
			logger.Logger.Error("[DATABASE-TRX-PANIC]", zap.Error(err))
		} else if err != nil {
			// Rollback if there was an error during the transaction
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("[DATABASE-ERROR] atomic err: %v, rollback err: %v", err, rbErr)
			}
			logger.Logger.Error("[DATABASE-TRX-ROLLBACK]", zap.Error(err))
		} else {
			// Commit if no errors occurred
			_ = tx.Commit()
		}
	}()

	err = fc(ctx, New(tx))
	if err != nil {
		tx.Error = err
	}

	return err
}
