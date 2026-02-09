package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TransactionTypeTransfer   TransactionType = "transfer"
	TransactionTypePayment    TransactionType = "payment"
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusReversed  TransactionStatus = "reversed"
)

type Transaction struct {
	ID            uuid.UUID         `json:"id"`
	FromAccountID uuid.UUID         `json:"from_account_id"`
	ToAccountID   uuid.UUID         `json:"to_account_id"`
	Amount        decimal.Decimal   `json:"amount"`
	Currency      string            `json:"currency"`
	Type          TransactionType   `json:"type"`
	Status        TransactionStatus `json:"status"`
	Description   string            `json:"description"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type TransferRequest struct {
	FromAccountID uuid.UUID       `json:"from_account_id"`
	ToAccountID   uuid.UUID       `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Description   string          `json:"description"`
}

// TransactionFilter holds query parameters for listing transactions.
type TransactionFilter struct {
	AccountID uuid.UUID
	Status    TransactionStatus
	Limit     int
	Offset    int
}

// TransactionEvent is published to NATS after a transaction completes.
type TransactionEvent struct {
	TransactionID uuid.UUID       `json:"transaction_id"`
	FromAccountID uuid.UUID       `json:"from_account_id"`
	ToAccountID   uuid.UUID       `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Type          TransactionType `json:"type"`
	Status        TransactionStatus `json:"status"`
	Timestamp     time.Time       `json:"timestamp"`
}
