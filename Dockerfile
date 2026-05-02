# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

# Install git for go mod download and ca-certificates for TLS
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency files first — Docker layer cache means these layers
# only rebuild when go.mod or go.sum change, not on every code change.
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy source
COPY . .

# Build a statically linked binary.
# -s -w strips debug info and DWARF tables → smaller binary.
# CGO_ENABLED=0 ensures no libc dependency → runs in scratch.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /app/bin/notes-upload-website \
    ./cmd/api

# ── Stage 2: Production image ─────────────────────────────────────────────────
FROM scratch

# Copy timezone data — needed for time.LoadLocation if you ever use it.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy TLS certificates — needed for outbound HTTPS (R2, Cloudflare).
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary.
COPY --from=builder /app/bin/notes-upload-website /notes-upload-website

# Copy migrations — the server runs these on startup.
COPY --from=builder /app/migrations /migrations

# Run as non-root for security.
# scratch has no useradd, so we use numeric UID directly.
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/notes-upload-website"]