package store

import (
	"sync"
	"time"

	"github.com/polai/backend/internal/models"
)

type MemoryStore struct {
	mu      sync.RWMutex
	markets map[uint64]*models.Market
	bets    map[string][]*models.Bet
	nextID  uint64
}

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		markets: make(map[uint64]*models.Market),
		bets:    make(map[string][]*models.Bet),
		nextID:  1,
	}
	s.seedInitialMarkets()
	return s
}

func (s *MemoryStore) seedInitialMarkets() {
	now := time.Now()
	s.markets[1] = &models.Market{
		ID:             1,
		Creator:        "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN",
		Resolver:       "GBJ66K3G44YF5YF5Z2KFXOYYPQQ4NGB5K7D2HQZ4DLY3O5ZJRPWZPO3V",
		Title:          "Will Stellar Soroban reach 50M total contract transactions in Q4?",
		Description:    "Resolves to YES if public ledger stats report 50,000,000+ contract invocations before Dec 31.",
		Category:       "Crypto",
		Deadline:       now.Add(14 * 24 * time.Hour),
		TotalYesAmount: 68000000000,
		TotalNoAmount:  32000000000,
		Outcome:        models.OutcomePending,
		CreatedAt:      now.Add(-2 * 24 * time.Hour),
	}
	s.markets[2] = &models.Market{
		ID:             2,
		Creator:        "GBJ66K3G44YF5YF5Z2KFXOYYPQQ4NGB5K7D2HQZ4DLY3O5ZJRPWZPO3V",
		Resolver:       "GBJ66K3G44YF5YF5Z2KFXOYYPQQ4NGB5K7D2HQZ4DLY3O5ZJRPWZPO3V",
		Title:          "Will OpenAI release GPT-5 before end of 2026?",
		Description:    "Resolves to YES upon public deployment of OpenAI model titled GPT-5.",
		Category:       "AI/Tech",
		Deadline:       now.Add(60 * 24 * time.Hour),
		TotalYesAmount: 42000000000,
		TotalNoAmount:  58000000000,
		Outcome:        models.OutcomePending,
		CreatedAt:      now.Add(-1 * 24 * time.Hour),
	}
	s.nextID = 3
}

func (s *MemoryStore) ListMarkets(category string) ([]*models.Market, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*models.Market, 0, len(s.markets))
	for _, m := range s.markets {
		if category == "" || m.Category == category {
			res = append(res, m)
		}
	}
	return res, nil
}

func (s *MemoryStore) GetMarket(id uint64) (*models.Market, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, ok := s.markets[id]
	if !ok {
		return nil, ErrNotFound
	}
	return m, nil
}

func (s *MemoryStore) CreateMarket(req models.CreateMarketRequest) (*models.Market, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m := &models.Market{
		ID:             s.nextID,
		Creator:        req.Creator,
		Resolver:       req.Resolver,
		Title:          req.Title,
		Description:    req.Description,
		Category:       req.Category,
		Deadline:       time.Unix(req.Deadline, 0),
		TotalYesAmount: 0,
		TotalNoAmount:  0,
		Outcome:        models.OutcomePending,
		CreatedAt:      time.Now(),
	}
	s.markets[s.nextID] = m
	s.nextID++
	return m, nil
}

func (s *MemoryStore) RecordBet(req models.PlaceBetRequest) (*models.Bet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, ok := s.markets[req.MarketID]
	if !ok {
		return nil, ErrNotFound
	}
	if time.Now().After(m.Deadline) || m.Outcome != models.OutcomePending {
		return nil, ErrAlreadyClosed
	}

	if req.Outcome == models.OutcomeYes {
		m.TotalYesAmount += req.Amount
	} else if req.Outcome == models.OutcomeNo {
		m.TotalNoAmount += req.Amount
	}

	bet := &models.Bet{
		ID:        req.TxHash,
		MarketID:  req.MarketID,
		User:      req.User,
		Outcome:   req.Outcome,
		Amount:    req.Amount,
		TxHash:    req.TxHash,
		Claimed:   false,
		CreatedAt: time.Now(),
	}

	s.bets[req.User] = append(s.bets[req.User], bet)
	return bet, nil
}

func (s *MemoryStore) GetUserBets(user string) ([]*models.Bet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.bets[user], nil
}
