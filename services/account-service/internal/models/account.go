package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type KYCStatus string

const (
	KYCPending  KYCStatus = "pending"
	KYCVerified KYCStatus = "verified"
	KYCRejected KYCStatus = "rejected"
)

type AccountType string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"
)

type AccountStatus string

const (
	AccountStatusActive AccountStatus = "active"
	AccountStatusFrozen AccountStatus = "frozen"
	AccountStatusClosed AccountStatus = "closed"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	KYCStatus KYCStatus `json:"kyc_status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Account struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	AccountType AccountType     `json:"account_type"`
	Balance     decimal.Decimal `json:"balance"`
	Currency    string          `json:"currency"`
	Status      AccountStatus   `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Request types

type CreateUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type CreateAccountRequest struct {
	UserID      string `json:"user_id"`
	AccountType string `json:"account_type"`
	Currency    string `json:"currency"`
}
