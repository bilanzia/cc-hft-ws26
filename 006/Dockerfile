# ==========================================
# Globale Versionen (Zentrale Pflege)
# ==========================================
ARG GO_VERSION=1.27
ARG GO_HASH=sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18

ARG ALPINE_VERSION=3.24
ARG ALPINE_HASH=sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

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
    go build -ldflags="-w -s" -o /nats-demo \
    main.go

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
COPY --from=builder /nats-demo ./nats-demo


# Ownership auf den neuen User übertragen
RUN chown appuser:appgroup ./nats-demo

# Ab hier läuft der Container als eingeschränkter Nutzer
USER appuser

EXPOSE 8080

CMD ["./nats-demo"]