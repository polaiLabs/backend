-- Migration for SQLite / Turso database
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
