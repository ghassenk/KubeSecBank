package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/shopspring/decimal"

	"github.com/ghassenk/KubeSecBank/services/transaction-service/internal/models"
	"github.com/ghassenk/KubeSecBank/services/transaction-service/internal/repository"
)

// balanceResponse is the expected shape from account-service's balance endpoint.
type balanceResponse struct {
	AccountID uuid.UUID       `json:"account_id"`
	Balance   decimal.Decimal `json:"balance"`
	Currency  string          `json:"currency"`
}

type TransactionHandler struct {
	repo              repository.TransactionRepository
	nc                *nats.Conn
	accountServiceURL string
}

func NewTransactionHandler(repo repository.TransactionRepository, nc *nats.Conn, accountServiceURL string) *TransactionHandler {
	return &TransactionHandler{
		repo:              repo,
		nc:                nc,
		accountServiceURL: accountServiceURL,
	}
}

// CreateTransfer handles POST /transactions/transfer
func (h *TransactionHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req models.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.FromAccountID == uuid.Nil || req.ToAccountID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "from_account_id and to_account_id are required")
		return
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if req.Currency == "" {
		writeError(w, http.StatusBadRequest, "currency is required")
		return
	}
	if req.FromAccountID == req.ToAccountID {
		writeError(w, http.StatusBadRequest, "cannot transfer to the same account")
		return
	}

	// Check balance via account-service
	balance, err := h.getBalance(req.FromAccountID, r.Header.Get("Authorization"))
	if err != nil {
		log.Printf("ERROR: check balance: %v", err)
		writeError(w, http.StatusBadGateway, "could not verify account balance")
		return
	}
	if balance.Balance.LessThan(req.Amount) {
		writeError(w, http.StatusUnprocessableEntity, "insufficient balance")
		return
	}

	// Create the transaction record
	now := time.Now().UTC()
	txn := &models.Transaction{
		ID:            uuid.New(),
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Type:          models.TransactionTypeTransfer,
		Status:        models.TransactionStatusPending,
		Description:   req.Description,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := h.repo.Create(r.Context(), txn); err != nil {
		log.Printf("ERROR: create transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "could not create transaction")
		return
	}

	// Mark as completed (in a real system this would involve a saga / 2PC)
	txn.Status = models.TransactionStatusCompleted
	txn.UpdatedAt = time.Now().UTC()
	if err := h.repo.UpdateStatus(r.Context(), txn.ID, models.TransactionStatusCompleted); err != nil {
		log.Printf("ERROR: update status: %v", err)
	}

	// Publish event to NATS
	event := models.TransactionEvent{
		TransactionID: txn.ID,
		FromAccountID: txn.FromAccountID,
		ToAccountID:   txn.ToAccountID,
		Amount:        txn.Amount,
		Currency:      txn.Currency,
		Type:          txn.Type,
		Status:        txn.Status,
		Timestamp:     txn.UpdatedAt,
	}
	eventData, err := json.Marshal(event)
	if err == nil {
		if err := h.nc.Publish("transactions.completed", eventData); err != nil {
			log.Printf("WARN: publish event: %v", err)
		}
	}

	writeJSON(w, http.StatusCreated, txn)
}

// GetTransaction handles GET /transactions/{id}
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	txn, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("ERROR: get transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if txn == nil {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}

	writeJSON(w, http.StatusOK, txn)
}

// ListTransactions handles GET /transactions?account_id=&status=&limit=&offset=
func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter := models.TransactionFilter{
		Limit:  20,
		Offset: 0,
	}

	if v := r.URL.Query().Get("account_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid account_id")
			return
		}
		filter.AccountID = id
	}

	if v := r.URL.Query().Get("status"); v != "" {
		filter.Status = models.TransactionStatus(v)
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			filter.Limit = n
		}
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			filter.Offset = n
		}
	}

	transactions, err := h.repo.List(r.Context(), filter)
	if err != nil {
		log.Printf("ERROR: list transactions: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
		"limit":        filter.Limit,
		"offset":       filter.Offset,
	})
}

// getBalance calls the account-service to retrieve the account balance.
func (h *TransactionHandler) getBalance(accountID uuid.UUID, authHeader string) (*balanceResponse, error) {
	url := fmt.Sprintf("%s/accounts/%s/balance", h.accountServiceURL, accountID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("account service request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("account service returned %d", resp.StatusCode)
	}

	var balance balanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&balance); err != nil {
		return nil, fmt.Errorf("decode balance: %w", err)
	}
	return &balance, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
