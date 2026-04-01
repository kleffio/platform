# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy module manifests first for better layer caching.
# GOWORK=off disables go.work so go.mod is used directly — all dependencies
# are fetched from the module proxy using their published versions.
COPY go.mod go.sum go.work go.work.sum ./
COPY packages/go.mod packages/

RUN GOWORK=off go mod download

# Copy full source and build.
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    GOWORK=off \
    go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /platform \
    ./cmd/api/

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3

RUN apk add --no-cache ca-certificates tzdata wget

RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /platform /platform

USER app

EXPOSE 8080

ENTRYPOINT ["/platform"]
