package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kubesec-bank/account-service/internal/models"
	"github.com/kubesec-bank/account-service/internal/repository"
)

type AccountHandler struct {
	repo repository.Repository
}

func NewAccountHandler(repo repository.Repository) *AccountHandler {
	return &AccountHandler{repo: repo}
}

func (h *AccountHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/users", h.handleUsers)
	mux.HandleFunc("/api/v1/users/", h.handleUserByID)
	mux.HandleFunc("/api/v1/accounts", h.handleAccounts)
	mux.HandleFunc("/api/v1/accounts/", h.handleAccountByID)
	mux.HandleFunc("/health", h.healthCheck)
}

func (h *AccountHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Users ---

func (h *AccountHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	h.CreateUser(w, r)
}

func (h *AccountHandler) handleUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	// Extract ID from /api/v1/users/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/users/"), "/")
	if parts[0] == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing user id"})
		return
	}

	// Check for nested accounts: /api/v1/users/{id}/accounts
	if len(parts) >= 2 && parts[1] == "accounts" {
		h.ListAccountsByUser(w, r, parts[0])
		return
	}

	h.GetUser(w, r, parts[0])
}

func (h *AccountHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.FullName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and full_name are required"})
		return
	}

	user := &models.User{
		Email:    req.Email,
		FullName: req.FullName,
	}

	if err := h.repo.CreateUser(r.Context(), user); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *AccountHandler) GetUser(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	user, err := h.repo.GetUser(r.Context(), id)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// --- Accounts ---

func (h *AccountHandler) handleAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	h.CreateAccount(w, r)
}

func (h *AccountHandler) handleAccountByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/accounts/")
	h.GetAccount(w, r, idStr)
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
		return
	}

	accountType := models.AccountType(req.AccountType)
	if accountType != models.AccountTypeChecking && accountType != models.AccountTypeSavings {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account_type must be checking or savings"})
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	account := &models.Account{
		UserID:      userID,
		AccountType: accountType,
		Currency:    currency,
	}

	if err := h.repo.CreateAccount(r.Context(), account); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
		return
	}

	writeJSON(w, http.StatusCreated, account)
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid account id"})
		return
	}

	account, err := h.repo.GetAccount(r.Context(), id)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "account not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get account"})
		return
	}

	writeJSON(w, http.StatusOK, account)
}

func (h *AccountHandler) ListAccountsByUser(w http.ResponseWriter, r *http.Request, userIDStr string) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	accounts, err := h.repo.ListAccountsByUser(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list accounts"})
		return
	}
	if accounts == nil {
		accounts = []models.Account{}
	}

	writeJSON(w, http.StatusOK, accounts)
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
