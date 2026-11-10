# ==========================================
# Globale Versionen (Zentrale Pflege)
# ==========================================
ARG GO_VERSION=1.27
ARG GO_HASH=sha256:d8c2c8f61535496a782e44d856d2a715f5c3574d538e1b65e90f230f6a27e7f9

ARG ALPINE_VERSION=3.20
ARG ALPINE_HASH=sha256:b89d9c93e9ed3597455c90a0b88a8bbb5cb7188438f70953fce22d6c84d0d10b

# ==========================================
# Build-Stage
# ==========================================
FROM golang:${GO_VERSION}-alpine@${GO_HASH} AS builder

# Setze Build-Umgebungsvariablen
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# Kopiere Dependency-Dateien (für besseres Layer-Caching)
COPY go.mod go.sum ./

# Nutze BuildKit-Cache für Go-Module
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Quellcode kopieren
COPY main.go ./

# Kompilieren mit Flags zur Größenoptimierung
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-w -s" -o /jobservice main.go

# ==========================================
# Run-Stage
# ==========================================
FROM alpine:${ALPINE_VERSION}@${ALPINE_HASH}

# Basis-Metadaten
LABEL org.opencontainers.image.title="NATS Job Service" \
      org.opencontainers.image.description="CloudEvents Producer/Consumer via NATS"

# Sicherheits-Best-Practice: Non-Root User anlegen
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Binary aus der Build-Stage kopieren
COPY --from=builder /build/jobservice ./jobservice

# Ownership auf den neuen User übertragen
RUN chown appuser:appgroup ./jobservice

# Ab hier läuft der Container als eingeschränkter Nutzer
USER appuser

EXPOSE 8080

CMD ["./jobservice"]