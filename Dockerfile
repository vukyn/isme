# isme — multi-stage build: UI (Vite) embedded into the Go binary via go:embed.
# The Go server serves the SPA from internal/web/dist (all:dist), so the UI MUST
# be built and placed there BEFORE `go build` — mirrors `make build-web`.

# ---- Stage 1: build the React/Vite UI ----
FROM node:20-alpine AS ui
WORKDIR /app/ui
# Cache deps on package manifests
COPY ui/package*.json ./
RUN npm install
COPY ui/ ./
RUN npm run build            # outputs ui/dist

# ---- Stage 2: build the Go binary (with embedded UI) ----
FROM golang:1.25 AS builder
WORKDIR /app
# Cache modules on go.mod/go.sum
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Place the built UI where go:embed (internal/web/dist) expects it
RUN rm -rf ./internal/web/dist
COPY --from=ui /app/ui/dist ./internal/web/dist
# Pure-Go build (modernc sqlite + bun pgdriver are CGO-free) → static binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/main ./cmd/main.go

# ---- Stage 3: minimal runtime ----
FROM scratch
# TLS roots for outbound HTTPS (isme → medioa avatar proxy, etc.)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Timezone data (scheduler / timestamptz handling)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/main /app/main
EXPOSE 8081
ENTRYPOINT ["/app/main"]
