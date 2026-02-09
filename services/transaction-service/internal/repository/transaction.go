package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/ghassenk/KubeSecBank/services/transaction-service/internal/models"
)

// TransactionRepository defines the data-access interface for transactions.
type TransactionRepository interface {
	Create(ctx context.Context, tx *models.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	List(ctx context.Context, filter models.TransactionFilter) ([]models.Transaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.TransactionStatus) error
}

// PostgresRepository implements TransactionRepository with PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create inserts a new transaction inside a database transaction for atomicity.
func (r *PostgresRepository) Create(ctx context.Context, txn *models.Transaction) error {
	dbTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer dbTx.Rollback()

	query := `
		INSERT INTO transactions (id, from_account_id, to_account_id, amount, currency, type, status, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = dbTx.ExecContext(ctx, query,
		txn.ID,
		txn.FromAccountID,
		txn.ToAccountID,
		txn.Amount,
		txn.Currency,
		txn.Type,
		txn.Status,
		txn.Description,
		txn.CreatedAt,
		txn.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}

	return dbTx.Commit()
}

// GetByID fetches a single transaction by its primary key.
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	query := `
		SELECT id, from_account_id, to_account_id, amount, currency, type, status, description, created_at, updated_at
		FROM transactions
		WHERE id = $1`

	var txn models.Transaction
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&txn.ID,
		&txn.FromAccountID,
		&txn.ToAccountID,
		&txn.Amount,
		&txn.Currency,
		&txn.Type,
		&txn.Status,
		&txn.Description,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w", err)
	}
	return &txn, nil
}

// List returns transactions matching the given filter with pagination.
func (r *PostgresRepository) List(ctx context.Context, filter models.TransactionFilter) ([]models.Transaction, error) {
	query := `
		SELECT id, from_account_id, to_account_id, amount, currency, type, status, description, created_at, updated_at
		FROM transactions
		WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if filter.AccountID != uuid.Nil {
		query += fmt.Sprintf(" AND (from_account_id = $%d OR to_account_id = $%d)", argIdx, argIdx+1)
		args = append(args, filter.AccountID, filter.AccountID)
		argIdx += 2
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
		argIdx++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var txn models.Transaction
		if err := rows.Scan(
			&txn.ID,
			&txn.FromAccountID,
			&txn.ToAccountID,
			&txn.Amount,
			&txn.Currency,
			&txn.Type,
			&txn.Status,
			&txn.Description,
			&txn.CreatedAt,
			&txn.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, txn)
	}
	return transactions, rows.Err()
}

// UpdateStatus changes the status of a transaction.
func (r *PostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.TransactionStatus) error {
	query := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("transaction %s not found", id)
	}
	return nil
}
