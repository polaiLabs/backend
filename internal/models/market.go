package models

import "time"

type Outcome string

const (
	OutcomePending Outcome = "PENDING"
	OutcomeYes     Outcome = "YES"
	OutcomeNo      Outcome = "NO"
	OutcomeInvalid Outcome = "INVALID"
)

type Market struct {
	ID             uint64    `json:"id"`
	Creator        string    `json:"creator"`
	Resolver       string    `json:"resolver"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Category       string    `json:"category"`
	Deadline       time.Time `json:"deadline"`
	TotalYesAmount int64     `json:"total_yes_amount"`
	TotalNoAmount  int64     `json:"total_no_amount"`
	Outcome        Outcome   `json:"outcome"`
	ResolvedAt     *time.Time`json:"resolved_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Bet struct {
	ID        string    `json:"id"`
	MarketID  uint64    `json:"market_id"`
	User      string    `json:"user"`
	Outcome   Outcome   `json:"outcome"`
	Amount    int64     `json:"amount"`
	TxHash    string    `json:"tx_hash"`
	Claimed   bool      `json:"claimed"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateMarketRequest struct {
	Creator     string `json:"creator"`
	Resolver    string `json:"resolver"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Deadline    int64  `json:"deadline"` // Unix timestamp in seconds
}

type PlaceBetRequest struct {
	MarketID uint64  `json:"market_id"`
	User     string  `json:"user"`
	Outcome  Outcome `json:"outcome"`
	Amount   int64   `json:"amount"`
	TxHash   string  `json:"tx_hash"`
}
