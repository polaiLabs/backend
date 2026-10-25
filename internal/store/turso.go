package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/polai/backend/internal/models"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type TursoStore struct {
	db *sql.DB
}

func NewTursoStore(databaseURL string, authToken string) (*TursoStore, error) {
	connStr := databaseURL
	if authToken != "" {
		connStr = fmt.Sprintf("%s?authToken=%s", databaseURL, authToken)
	}

	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping Turso database: %w", err)
	}

	s := &TursoStore{db: db}
	if err := s.autoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to auto migrate schema: %w", err)
	}

	return s, nil
}

func (s *TursoStore) autoMigrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS markets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		creator TEXT NOT NULL,
		resolver TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		category TEXT NOT NULL,
		deadline DATETIME NOT NULL,
		total_yes_amount INTEGER DEFAULT 0,
		total_no_amount INTEGER DEFAULT 0,
		outcome TEXT DEFAULT 'PENDING',
		resolved_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS bets (
		id TEXT PRIMARY KEY,
		market_id INTEGER NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
		user_address TEXT NOT NULL,
		outcome TEXT NOT NULL,
		amount INTEGER NOT NULL,
		tx_hash TEXT NOT NULL,
		claimed BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_bets_user ON bets(user_address);
	CREATE INDEX IF NOT EXISTS idx_bets_market ON bets(market_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *TursoStore) ListMarkets(category string) ([]*models.Market, error) {
	query := `SELECT id, creator, resolver, title, description, category, deadline, 
	          total_yes_amount, total_no_amount, outcome, resolved_at, created_at 
	          FROM markets`
	var rows *sql.Rows
	var err error

	if category != "" {
		query += " WHERE category = ? ORDER BY id DESC"
		rows, err = s.db.Query(query, category)
	} else {
		query += " ORDER BY id DESC"
		rows, err = s.db.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []*models.Market
	for rows.Next() {
		var m models.Market
		var deadlineStr, createdAtStr string
		var resolvedAtStr sql.NullString

		err := rows.Scan(
			&m.ID, &m.Creator, &m.Resolver, &m.Title, &m.Description, &m.Category,
			&deadlineStr, &m.TotalYesAmount, &m.TotalNoAmount, &m.Outcome,
			&resolvedAtStr, &createdAtStr,
		)
		if err != nil {
			return nil, err
		}

		m.Deadline, _ = time.Parse(time.RFC3339, deadlineStr)
		m.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		if resolvedAtStr.Valid {
			t, _ := time.Parse(time.RFC3339, resolvedAtStr.String)
			m.ResolvedAt = &t
		}

		markets = append(markets, &m)
	}

	return markets, nil
}

func (s *TursoStore) GetMarket(id uint64) (*models.Market, error) {
	query := `SELECT id, creator, resolver, title, description, category, deadline, 
	          total_yes_amount, total_no_amount, outcome, resolved_at, created_at 
	          FROM markets WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var m models.Market
	var deadlineStr, createdAtStr string
	var resolvedAtStr sql.NullString

	err := row.Scan(
		&m.ID, &m.Creator, &m.Resolver, &m.Title, &m.Description, &m.Category,
		&deadlineStr, &m.TotalYesAmount, &m.TotalNoAmount, &m.Outcome,
		&resolvedAtStr, &createdAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	m.Deadline, _ = time.Parse(time.RFC3339, deadlineStr)
	m.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	if resolvedAtStr.Valid {
		t, _ := time.Parse(time.RFC3339, resolvedAtStr.String)
		m.ResolvedAt = &t
	}

	return &m, nil
}

func (s *TursoStore) CreateMarket(req models.CreateMarketRequest) (*models.Market, error) {
	deadline := time.Unix(req.Deadline, 0)
	now := time.Now()

	query := `INSERT INTO markets (creator, resolver, title, description, category, deadline, total_yes_amount, total_no_amount, outcome, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, 0, 0, 'PENDING', ?) RETURNING id`

	var id uint64
	err := s.db.QueryRow(query, req.Creator, req.Resolver, req.Title, req.Description, req.Category, deadline.Format(time.RFC3339), now.Format(time.RFC3339)).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &models.Market{
		ID:             id,
		Creator:        req.Creator,
		Resolver:       req.Resolver,
		Title:          req.Title,
		Description:    req.Description,
		Category:       req.Category,
		Deadline:       deadline,
		TotalYesAmount: 0,
		TotalNoAmount:  0,
		Outcome:        models.OutcomePending,
		CreatedAt:      now,
	}, nil
}

func (s *TursoStore) RecordBet(req models.PlaceBetRequest) (*models.Bet, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check market state
	var outcome string
	var deadlineStr string
	err = tx.QueryRow(`SELECT outcome, deadline FROM markets WHERE id = ?`, req.MarketID).Scan(&outcome, &deadlineStr)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if outcome != "PENDING" {
		return nil, ErrAlreadyClosed
	}

	// Update pool amounts
	if req.Outcome == models.OutcomeYes {
		_, err = tx.Exec(`UPDATE markets SET total_yes_amount = total_yes_amount + ? WHERE id = ?`, req.Amount, req.MarketID)
	} else if req.Outcome == models.OutcomeNo {
		_, err = tx.Exec(`UPDATE markets SET total_no_amount = total_no_amount + ? WHERE id = ?`, req.Amount, req.MarketID)
	}
	if err != nil {
		return nil, err
	}

	// Insert bet
	now := time.Now()
	_, err = tx.Exec(
		`INSERT INTO bets (id, market_id, user_address, outcome, amount, tx_hash, claimed, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, 0, ?)`,
		req.TxHash, req.MarketID, req.User, string(req.Outcome), req.Amount, req.TxHash, now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Bet{
		ID:        req.TxHash,
		MarketID:  req.MarketID,
		User:      req.User,
		Outcome:   req.Outcome,
		Amount:    req.Amount,
		TxHash:    req.TxHash,
		Claimed:   false,
		CreatedAt: now,
	}, nil
}

func (s *TursoStore) GetUserBets(user string) ([]*models.Bet, error) {
	query := `SELECT id, market_id, user_address, outcome, amount, tx_hash, claimed, created_at
	          FROM bets WHERE user_address = ? ORDER BY created_at DESC`
	rows, err := s.db.Query(query, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bets []*models.Bet
	for rows.Next() {
		var b models.Bet
		var outcomeStr, createdAtStr string
		err := rows.Scan(
			&b.ID, &b.MarketID, &b.User, &outcomeStr, &b.Amount, &b.TxHash, &b.Claimed, &createdAtStr,
		)
		if err != nil {
			return nil, err
		}
		b.Outcome = models.Outcome(outcomeStr)
		b.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		bets = append(bets, &b)
	}

	return bets, nil
}
