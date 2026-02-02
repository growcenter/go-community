package pgsql

import (
	"context"
	"fmt"
	"go-community/internal/pkg/logger"

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

	logger.EnrichContextMap(context.Background(), map[string]interface{}{
		"type":    "TransactionBegin",
		"code":    "DATABASE_TRANSACTION_BEGIN",
		"message": "[DATABASE-TRX-BEGIN]",
	})

	defer func() {
		if r := recover(); r != nil {
			dtx.Rollback()
			err := fmt.Errorf("[DATABASE-ERROR] panic happened because: " + fmt.Sprintf("%v", r))
			logger.LogError(context.Background(), logger.ErrorContext{
				Type:      "PanicError",
				Code:      "DATABASE_TRANSACTION_PANIC",
				Message:   err.Error(),
				Retriable: false,
			})
		}
	}()

	if err := fc(dtx); err != nil {
		dtx.Rollback()
		logger.LogError(context.Background(), logger.ErrorContext{
			Type:      "RollbackError",
			Code:      "DATABASE_TRANSACTION_ROLLBACK",
			Message:   err.Error(),
			Retriable: false,
		})
		return err
	}

	return dtx.Commit().Error
}

// Atomic executes the provided function within a database transaction.
// This is the recommended way to handle transactions as it ensures proper cleanup.
//
// Transaction Lifecycle:
// 1. Begin transaction
// 2. Execute the provided function
// 3. If function succeeds: Commit
// 4. If function fails or panics: Rollback
//
// Example usage:
//
//	err := repo.Transaction.Atomic(ctx, func(ctx context.Context, r *PostgreRepositories) error {
//	    // Perform database operations using r
//	    if err := r.SomeRepo.Create(ctx, data); err != nil {
//	        return err // This will trigger a rollback
//	    }
//	    return nil // This will trigger a commit
//	})
func (tr *transactionRepository) Atomic(ctx context.Context, fc func(ctx context.Context, r *PostgreRepositories) error) error {
	// Step 1: Begin the transaction with context support
	// WithContext ensures the transaction respects context cancellation
	tx := tr.db.WithContext(ctx).Begin()
	err := tx.Error
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Step 2: Set up deferred cleanup to ensure transaction is always finalized
	// This runs after the function completes, regardless of success or failure
	defer func() {
		// Step 2a: Handle panic recovery
		// If a panic occurs during the transaction, we catch it here
		if p := recover(); p != nil {
			// Rollback the transaction due to panic
			tx.Rollback()
			// Convert panic to error for consistent error handling
			err = fmt.Errorf("transaction failed due to panic: %v", p)

			// Log the panic with rich error context
			// Note: The updated context from LogError won't propagate from defer,
			// but the middleware will still pick it up from the original context
			logger.LogError(ctx, logger.ErrorContext{
				Type:      "PanicError",
				Code:      "DATABASE_TRANSACTION_PANIC",
				Message:   err.Error(),
				Retriable: false,
			})

		} else if err != nil {
			// Step 2b: Handle explicit errors returned by the function
			// If the function returned an error, rollback the transaction
			if rbErr := tx.Rollback().Error; rbErr != nil {
				// If rollback itself fails, wrap both errors for debugging
				err = fmt.Errorf("failed to rollback transaction (original error: %w): %v", err, rbErr)

				logger.LogError(ctx, logger.ErrorContext{
					Type:      "RollbackError",
					Code:      "DATABASE_TRANSACTION_ROLLBACK",
					Message:   err.Error(),
					Retriable: false,
				})
			}
		} else {
			// Step 2c: Success path - commit the transaction
			// If no error occurred, commit all changes
			if commitErr := tx.Commit().Error; commitErr != nil {
				// If commit fails, update err so it's returned to caller
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)

				logger.LogError(ctx, logger.ErrorContext{
					Type:      "CommitError",
					Code:      "DATABASE_TRANSACTION_COMMIT",
					Message:   err.Error(),
					Retriable: false,
				})
			}
		}
	}()

	// Step 3: Execute the user-provided function with a new repository instance
	// The repository uses the transaction (tx) instead of the main DB connection
	// This ensures all operations within the function are part of the same transaction
	err = fc(ctx, New(tx))
	if err != nil {
		// Error will be handled by the defer block above (Step 2b)
		return err
	}

	// Step 4: Return nil on success
	// The defer block will commit the transaction (Step 2c)
	return nil
}
