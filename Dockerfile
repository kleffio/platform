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
# distroless/static has no shell, package manager, or libc — minimal attack surface.
FROM gcr.io/distroless/static-debian12:nonroot

# CA certificates and timezone data from the builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /platform /platform

EXPOSE 8080

ENTRYPOINT ["/platform"]
