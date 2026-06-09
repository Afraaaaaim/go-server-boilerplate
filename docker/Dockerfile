# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

# Install git for go mod download (some modules need it)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build a statically linked binary
# -s strips debug symbols, -w strips DWARF info → smaller binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o server ./cmd/api

# ─── Stage 2: Runtime ─────────────────────────────────────────────────────────
# distroless has no shell, no package manager, minimal attack surface
FROM gcr.io/distroless/static-debian12

# Copy timezone data and certs from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /build/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]