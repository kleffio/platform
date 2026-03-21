# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy workspace and module manifests first for better layer caching.
# The workspace includes the root module and ./packages (go-common).
COPY go.work ./
COPY go.mod ./
COPY packages/go.mod packages/

# Sync the workspace so Go resolves all local module replacements.
RUN go work sync

# Copy the full source tree and build the production binary.
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /platform \
    ./cmd/api/

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
# alpine provides a minimal shell and wget so Docker health checks work.
FROM alpine:3

RUN apk add --no-cache ca-certificates tzdata wget

RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /platform /platform

USER app

EXPOSE 8080

ENTRYPOINT ["/platform"]
