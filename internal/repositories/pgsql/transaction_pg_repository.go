package pgsql

import (
	"context"
	"fmt"
	"go-community/internal/pkg/logger"

	"gorm.io/gorm"
)

// txKey is an unexported type used as a context key for the transactional *gorm.DB.
// Using a dedicated type (instead of a plain string) prevents collisions with other
// packages that also store values in context.
type txKey struct{}

// TxFromContext returns the transactional *gorm.DB stored by Atomic, and whether one
// was found. Repositories call this to transparently participate in the active transaction.
func TxFromContext(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	return tx, ok
}

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

	logger.Add(context.Background(), "transaction_begin", true)

	defer func() {
		if r := recover(); r != nil {
			dtx.Rollback()
			err := fmt.Errorf("[DATABASE-ERROR] panic happened because: " + fmt.Sprintf("%v", r))
			logger.AddError(context.Background(), &logger.ErrorContext{
				Type:      "PanicError",
				Code:      "DATABASE_TRANSACTION_PANIC",
				Message:   err.Error(),
				Retriable: false,
			})
		}
	}()

	if err := fc(dtx); err != nil {
		dtx.Rollback()
		logger.AddError(context.Background(), &logger.ErrorContext{
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
func (tr *transactionRepository) Atomic(ctx context.Context, fc func(ctx context.Context, r *PostgreRepositories) error) (atomicError error) {
	// Step 1: Begin the transaction with context support
	// WithContext ensures the transaction respects context cancellation
	tx := tr.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Step 2: Set up deferred cleanup to ensure transaction is always finalized.
	// IMPORTANT: atomicError is a named return so the defer can overwrite it even when
	// the function exits via panic recovery (without named returns, recover() sets
	// a local variable but the caller always sees nil — the zero value of error).
	defer func() {
		// Step 2a: Handle panic recovery
		// If a panic occurs during the transaction, we catch it here
		if p := recover(); p != nil {
			// Rollback the transaction due to panic
			tx.Rollback()
			// Convert panic to error for consistent error handling and propagate
			// it back to the caller via the named return.
			atomicError = fmt.Errorf("transaction failed due to panic: %v", p)

			// Log the panic with rich error context
			logger.AddError(ctx, &logger.ErrorContext{
				Type:      "PanicError",
				Code:      "DATABASE_TRANSACTION_PANIC",
				Message:   atomicError.Error(),
				Retriable: false,
			})

		} else if atomicError != nil {
			// Step 2b: Handle explicit errors returned by the function
			// If the function returned an error, rollback the transaction
			if rbErr := tx.Rollback().Error; rbErr != nil {
				// If rollback itself fails, wrap both errors for debugging
				atomicError = fmt.Errorf("failed to rollback transaction (original error: %w): %v", atomicError, rbErr)

				logger.AddError(ctx, &logger.ErrorContext{
					Type:      "RollbackError",
					Code:      "DATABASE_TRANSACTION_ROLLBACK",
					Message:   atomicError.Error(),
					Retriable: false,
				})
			}
		} else {
			// Step 2c: Success path - commit the transaction
			// If no error occurred, commit all changes
			if commitErr := tx.Commit().Error; commitErr != nil {
				// If commit fails, update atomicError so it's returned to caller
				atomicError = fmt.Errorf("failed to commit transaction: %w", commitErr)

				logger.AddError(ctx, &logger.ErrorContext{
					Type:      "CommitError",
					Code:      "DATABASE_TRANSACTION_COMMIT",
					Message:   atomicError.Error(),
					Retriable: false,
				})
			}
		}
	}()

	// Step 3: Execute the user-provided function with a new repository instance.
	// Embed the transactional DB in context so repositories can pick it up without
	// needing the caller to explicitly pass *PostgreRepositories through every call.
	ctx = context.WithValue(ctx, txKey{}, tx)
	atomicError = fc(ctx, New(tx))

	// Step 4: Return atomicError (nil on success, non-nil on fc failure).
	// The defer block finalises the transaction (commit or rollback).
	return atomicError
}
