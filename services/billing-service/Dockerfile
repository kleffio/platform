# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go-common shared module (if imported by this service)
COPY packages/go-common/ ./packages/go-common/

# Copy service source
COPY services/billing-service/ ./services/billing-service/

WORKDIR /app/services/billing-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api

# ─── Stage 2: Run ─────────────────────────────────────────────────────────────
FROM alpine:3.21 AS runner

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/api ./api

EXPOSE 8080

CMD ["./api"]
