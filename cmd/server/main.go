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

	databaseURL := os.Getenv("DATABASE_URL")
	tursoAuthToken := os.Getenv("TURSO_AUTH_TOKEN")

	var appStore store.Store

	if databaseURL != "" {
		log.Printf("Connecting to Turso / libSQL SQLite database at: %s", databaseURL)
		tStore, err := store.NewTursoStore(databaseURL, tursoAuthToken)
		if err != nil {
			log.Fatalf("Failed to initialize Turso database: %v", err)
		}
		appStore = tStore
		log.Println("Turso SQLite store initialized and schema verified successfully.")
	} else {
		log.Println("No DATABASE_URL configured. Falling back to in-memory store.")
		appStore = store.NewMemoryStore()
	}

	api := handlers.NewAPIHandler(appStore)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"polai-backend","db_driver":"libsql"}`))
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
