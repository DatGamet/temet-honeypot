# ---- Build Stage ----
FROM golang:1.26.4-alpine AS builder

WORKDIR /app

# Abhängigkeiten für Build (z.B. git, falls private Module benötigt werden)
RUN apk add --no-cache git ca-certificates

# go.mod und go.sum zuerst kopieren, um Layer-Caching zu nutzen
COPY go.mod go.sum ./
RUN go mod download

# Restlichen Quellcode kopieren
COPY . .

# Statisch kompilieren (CGO_ENABLED=0) für ein minimales, plattformunabhängiges Binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bot .

# ---- Final Stage ----
FROM alpine:3.20

WORKDIR /app

# CA-Zertifikate für HTTPS-Verbindungen zur Discord API
RUN apk add --no-cache ca-certificates tzdata

# Nicht-root User anlegen
RUN addgroup -S botgroup && adduser -S botuser -G botgroup

COPY --from=builder /app/bot .

USER botuser

ENTRYPOINT ["./bot"]