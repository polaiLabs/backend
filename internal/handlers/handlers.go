package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/polai/backend/internal/models"
	"github.com/polai/backend/internal/store"
)

type APIHandler struct {
	store store.Store
}

func NewAPIHandler(s store.Store) *APIHandler {
	return &APIHandler{store: s}
}

func (h *APIHandler) ListMarkets(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	markets, err := h.store.ListMarkets(category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": markets,
	})
}

func (h *APIHandler) GetMarket(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/markets/")
	id, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid market id", http.StatusBadRequest)
		return
	}

	m, err := h.store.GetMarket(id)
	if err != nil {
		http.Error(w, "market not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": m,
	})
}

func (h *APIHandler) CreateMarket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateMarketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	m, err := h.store.CreateMarket(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": m,
	})
}

func (h *APIHandler) RecordBet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.PlaceBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	bet, err := h.store.RecordBet(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": bet,
	})
}

func (h *APIHandler) GetUserPositions(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "missing user address parameter", http.StatusBadRequest)
		return
	}

	bets, err := h.store.GetUserBets(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": bets,
	})
}

func WithCORS(next http.Handler) http.Handler {
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv == "" {
		allowedOriginsEnv = "*"
	}
	allowedList := strings.Split(allowedOriginsEnv, ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowOrigin := "*"

		if allowedOriginsEnv != "*" && origin != "" {
			for _, allowed := range allowedList {
				trimmed := strings.TrimSpace(allowed)
				// Match exact, or with/without protocol scheme
				if trimmed == origin || 
				   "https://" + trimmed == origin || 
				   "http://" + trimmed == origin ||
				   trimmed == "*" {
					allowOrigin = origin
					break
				}
			}
		} else if allowedOriginsEnv != "*" {
			// Fallback to first configured origin if no origin header
			allowOrigin = strings.TrimSpace(allowedList[0])
		}

		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
