# Multi-stage Dockerfile for high-performance minimal container
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency manifests
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build statically compiled binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Final lightweight scratch/alpine runtime
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server /app/server

# Port is provided dynamically by Render via $PORT environment variable (defaults to 8080)
EXPOSE 8080

CMD ["/app/server"]
