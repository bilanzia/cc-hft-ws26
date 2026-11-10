# NATS Demo

Dieser Service demonstriert eine moderne ereignisgesteuerte Architektur (Event-Driven Architecture) in **Go 1.27**. Er implementiert das Producer/Consumer-Pattern über **NATS** unter strikter Verwendung der **CloudEvents (CNCF)** Spezifikation.

## Features

* **Single Binary, Multi-Mode:** Die Anwendung wird als ein einziges Image gebaut und kann über die Umgebungsvariable `APP_MODE` entweder als HTTP-Producer oder als NATS-Consumer gestartet werden.
* **CloudEvents Native:** Volle Integration des offiziellen CloudEvents Go-SDKs für standardisierte Nachrichten-Payloads.
* **High-Performance HTTP:** Routing über den leichtgewichtigen `go-chi/chi` Router.
* **Docker Gold Standard:** Das Container-Image nutzt Multi-Stage-Builds, BuildKit-Caching, Non-Root-Execution und striktes Hash-Pinning (Zero-Trust) für maximale Sicherheit und Reproduzierbarkeit.

---

## Voraussetzungen

* Docker & Docker Compose installiert
* (Optional) Go 1.27+ für lokale Entwicklung
* (Optional) [Swaggo CLI](https://github.com/swaggo/swag) zur Aktualisierung der OpenAPI-Dokumentation

---

## Schnellstart

Starten der gesamten Infrastruktur (NATS Broker, 1x Producer, 1x Consumer):

```bash
docker-compose up --build
```

Der HTTP-Router des Producers lauscht anschließend auf `http://localhost:8080`.

---

## API Nutzung & Testen

Der Producer stellt einen REST-Endpunkt zur Verfügung, der eingehende Jobs in CloudEvents umwandelt und an das NATS-Subject `jobs.notify` sendet.

### 1. Erfolgreicher Durchlauf (Channel: "email")
Der Consumer filtert aktiv nach dem Kanal `email`. Wenn dieser vorhanden ist, wird die Verarbeitung geloggt.

```bash
curl -X POST http://localhost:8080/v1/jobs:notify \
  -H "Content-Type: application/json" \
  -d '{
        "job_id": "job-123", 
        "channels": ["sms", "email"]
      }'
```
*Erwarteter Consumer-Log:* `successfully send to channel: email (JobID: job-123...)`

### 2. Gefilterter Durchlauf
Fehlt der `email`-Kanal in der Payload, ignoriert der Consumer das Event.

```bash
curl -X POST http://localhost:8080/v1/jobs:notify \
  -H "Content-Type: application/json" \
  -d '{
        "job_id": "job-456", 
        "channels": ["push"]
      }'
```
*Erwarteter Consumer-Log:* `Ignoriere Event ... - kein 'email' Channel`

---

## OpenAPI / Swagger Dokumentation

Die API ist mittels Swaggo dokumentiert. Um die Spezifikation (Swagger 2.0 / OpenAPI 3.0) nach einer Code-Änderung neu zu generieren, führe lokal folgenden Befehl aus:

```bash
swag init
```

Die generierten Dateien (`swagger.json` und `swagger.yaml`) befinden sich im Ordner `/docs` und können in Tools wie Postman oder Swagger UI importiert werden.
