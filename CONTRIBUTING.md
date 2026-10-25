# Contributing to Polai Backend

Thank you for your interest in contributing to **Polai**! This repository powers the indexing, caching, and REST API layer for the Polai Prediction Market.

## Getting Started

### Prerequisites
- [Go](https://go.dev/doc/install) (v1.22+)

### Local Development & Testing
```bash
# Clone the repository
git clone https://github.com/polaiLabs/backend.git
cd backend

# Run tests
go test -v ./...

# Start server locally
go run cmd/server/main.go
```

## Drips Wave Contributions
We welcome contributors from the **Drips Wave** program!

- Look for issues tagged with `Stellar Wave`.
- Apply to issues via the Drips Wave portal before opening a PR.
- Write unit tests for new store methods or handlers.
- Verify code formatting with `go fmt ./...`.
