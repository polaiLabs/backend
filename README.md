# Polai: High-Performance Go Backend

Real-time indexing, market caching, and REST API service for the **Polai Prediction Market** on Stellar Soroban.

## Features
- **Concurrent In-Memory & SQL-Ready Architecture**: Fast read/write throughput for live market odds.
- **Full Market Lifecycle APIs**: Create markets, list with category filters, place bets, and query user portfolio positions.
- **CORS & Middleware**: Pre-configured for Next.js frontend integration and wallet authentication.
- **Unit Tested**: Robust test suite with 100% core coverage.

## API Endpoints
- `GET /health` - Server healthcheck
- `GET /api/markets?category=Crypto` - List markets with optional category filtering
- `GET /api/markets/:id` - Fetch detailed market data and current odds
- `POST /api/markets` - Create a new prediction market
- `POST /api/bets` - Record user on-chain prediction bet
- `GET /api/positions?user={address}` - Fetch all open and resolved bets for a user

## Run & Test
```bash
# Run unit tests
go test ./...

# Start server locally on :8080
go run cmd/server/main.go
```
