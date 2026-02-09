package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/kubesec-bank/account-service/internal/models"
	"github.com/shopspring/decimal"
)

type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)

	CreateAccount(ctx context.Context, account *models.Account) error
	GetAccount(ctx context.Context, id uuid.UUID) (*models.Account, error)
	ListAccountsByUser(ctx context.Context, userID uuid.UUID) ([]models.Account, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	user.ID = uuid.New()
	user.KYCStatus = models.KYCPending
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = user.CreatedAt

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, full_name, kyc_status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email, user.FullName, user.KYCStatus, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, full_name, kyc_status, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Email, &user.FullName, &user.KYCStatus, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresRepository) CreateAccount(ctx context.Context, account *models.Account) error {
	account.ID = uuid.New()
	account.Balance = decimal.Zero
	account.Status = models.AccountStatusActive
	account.CreatedAt = time.Now().UTC()
	account.UpdatedAt = account.CreatedAt

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts (id, user_id, account_type, balance, currency, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		account.ID, account.UserID, account.AccountType, account.Balance.String(),
		account.Currency, account.Status, account.CreatedAt, account.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetAccount(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	account := &models.Account{}
	var balanceStr string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, account_type, balance, currency, status, created_at, updated_at
		 FROM accounts WHERE id = $1`, id,
	).Scan(&account.ID, &account.UserID, &account.AccountType, &balanceStr,
		&account.Currency, &account.Status, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		return nil, err
	}
	account.Balance, err = decimal.NewFromString(balanceStr)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *PostgresRepository) ListAccountsByUser(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, account_type, balance, currency, status, created_at, updated_at
		 FROM accounts WHERE user_id = $1 ORDER BY created_at`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		var balanceStr string
		if err := rows.Scan(&a.ID, &a.UserID, &a.AccountType, &balanceStr,
			&a.Currency, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.Balance, err = decimal.NewFromString(balanceStr)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}
