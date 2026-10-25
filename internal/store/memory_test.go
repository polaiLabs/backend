package store

import (
	"testing"
	"time"

	"github.com/polai/backend/internal/models"
)

func TestMemoryStore_Lifecycle(t *testing.T) {
	s := NewMemoryStore()

	// Initial seeds
	markets, err := s.ListMarkets("")
	if err != nil {
		t.Fatalf("failed to list markets: %v", err)
	}
	if len(markets) != 2 {
		t.Fatalf("expected 2 seed markets, got %d", len(markets))
	}

	// Create new market
	newMkt, err := s.CreateMarket(models.CreateMarketRequest{
		Creator:     "GCreator",
		Resolver:    "GResolver",
		Title:       "Will Soroban TVL cross 100M?",
		Description: "TVL metric test",
		Category:    "Crypto",
		Deadline:    time.Now().Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("failed to create market: %v", err)
	}

	if newMkt.ID != 3 {
		t.Errorf("expected new market ID 3, got %d", newMkt.ID)
	}

	// Place Bet
	bet, err := s.RecordBet(models.PlaceBetRequest{
		MarketID: 3,
		User:     "GUser1",
		Outcome:  models.OutcomeYes,
		Amount:   50000000,
		TxHash:   "0xabc123",
	})
	if err != nil {
		t.Fatalf("failed to record bet: %v", err)
	}
	if bet.Amount != 50000000 {
		t.Errorf("expected bet amount 50000000, got %d", bet.Amount)
	}

	// Check updated market
	m, err := s.GetMarket(3)
	if err != nil {
		t.Fatalf("failed to get market 3: %v", err)
	}
	if m.TotalYesAmount != 50000000 {
		t.Errorf("expected total yes 50000000, got %d", m.TotalYesAmount)
	}

	// Check user positions
	userBets, err := s.GetUserBets("GUser1")
	if err != nil {
		t.Fatalf("failed to get user bets: %v", err)
	}
	if len(userBets) != 1 {
		t.Errorf("expected 1 bet for user, got %d", len(userBets))
	}
}
