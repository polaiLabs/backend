package main

import (
	"log"
	"net/http"
	"os"

	"github.com/polai/backend/internal/handlers"
	"github.com/polai/backend/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	memStore := store.NewMemoryStore()
	api := handlers.NewAPIHandler(memStore)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"polai-backend"}`))
	})

	// Market endpoints
	mux.HandleFunc("/api/markets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.CreateMarket(w, r)
		} else {
			api.ListMarkets(w, r)
		}
	})
	mux.HandleFunc("/api/markets/", api.GetMarket)
	mux.HandleFunc("/api/bets", api.RecordBet)
	mux.HandleFunc("/api/positions", api.GetUserPositions)

	handler := handlers.WithCORS(mux)

	log.Printf("Polai Backend server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
