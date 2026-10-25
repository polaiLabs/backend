package store

import (
	"errors"

	"github.com/polai/backend/internal/models"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyClosed = errors.New("market has already closed")
)

type Store interface {
	ListMarkets(category string) ([]*models.Market, error)
	GetMarket(id uint64) (*models.Market, error)
	CreateMarket(req models.CreateMarketRequest) (*models.Market, error)
	RecordBet(req models.PlaceBetRequest) (*models.Bet, error)
	GetUserBets(user string) ([]*models.Bet, error)
}
