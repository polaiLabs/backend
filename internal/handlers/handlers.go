package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/polai/backend/internal/models"
	"github.com/polai/backend/internal/store"
)

type APIHandler struct {
	store *store.MemoryStore
}

func NewAPIHandler(s *store.MemoryStore) *APIHandler {
	return &APIHandler{store: s}
}

func (h *APIHandler) ListMarkets(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	markets := h.store.ListMarkets(category)

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

	bets := h.store.GetUserBets(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": bets,
	})
}

func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
