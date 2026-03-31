# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy module manifests first for better layer caching.
COPY plugins/plugin-sdk-go/go.mod plugins/plugin-sdk-go/go.sum plugins/plugin-sdk-go/
COPY platform/go.work platform/go.work.sum platform/
COPY platform/go.mod platform/go.sum platform/
COPY platform/packages/go.mod platform/packages/

WORKDIR /build/platform
RUN go work sync

# Copy full source and build.
WORKDIR /build
COPY plugins/plugin-sdk-go/ plugins/plugin-sdk-go/
COPY platform/ platform/

WORKDIR /build/platform
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
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
