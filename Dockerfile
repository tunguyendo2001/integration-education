FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o ingestion-worker ./cmd/ingestion-worker
RUN CGO_ENABLED=0 GOOS=linux go build -o sync-worker ./cmd/sync-worker
RUN CGO_ENABLED=0 GOOS=linux go build -o pull-worker ./cmd/pull-worker

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/api .
COPY --from=builder /app/ingestion-worker .
COPY --from=builder /app/sync-worker .
COPY --from=builder /app/pull-worker .

# Copy config
COPY config.yaml .

CMD ["./api"]
