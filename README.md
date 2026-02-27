# Global Supply Shock Early-Warning Platform

A highly scalable event-driven platform written in Go to detect and surface disruption risks for essential goods across countries.

## What It Solves

Governments, NGOs, and distributors need early warning when medicines, food staples, or fuel are at risk due to shipping delays, weather disruption, port congestion, price spikes, or geopolitical events.

This platform ingests global signals, computes risk scores by geography and commodity, and raises actionable alerts.

## Architecture (MVP)

Services:

- `ingest`: receives external signal events and publishes to Kafka topic `signals.raw`
- `risk-engine`: consumes raw signals, computes risk score, stores to Postgres, publishes to `risk.scored`
- `alert-service`: consumes risk events and opens alerts based on threshold + cooldown policies
- `query-api`: provides read/update APIs for risks, alerts, and dashboard summary

Infrastructure:

- Kafka (Redpanda in local compose)
- PostgreSQL
- Docker Compose for local end-to-end runtime

## Repository Layout

- `cmd/` - service entrypoints
- `internal/contracts` - shared event contracts
- `internal/risk` - scoring engine
- `internal/storage` - Postgres access + embedded SQL migrations
- `deploy/k8s` - Kubernetes starter manifests
- `migrations` - SQL migrations mirror
- `docs` - architecture, API, deployment, scaling notes

## Quick Start (Local)

1. Start stack:

```bash
docker compose up --build
```

2. Check health:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8081/healthz
```

3. Publish simulated events:

```bash
curl -X POST http://localhost:8081/v1/simulate -H "Content-Type: application/json" -d '{"count": 200}'
```

4. Query risk and alerts:

```bash
curl "http://localhost:8080/v1/risks?limit=20"
curl "http://localhost:8080/v1/alerts?status=open&limit=20"
curl "http://localhost:8080/v1/dashboard/summary"
```

## Manual Run (without Docker)

Prereqs:

- Go 1.26+
- Postgres running
- Kafka-compatible broker running

Then:

```bash
go mod tidy
go run ./cmd/query-api
go run ./cmd/risk-engine
go run ./cmd/alert-service
go run ./cmd/ingest
```

## API Summary

- `POST /v1/signals` (ingest)
- `POST /v1/simulate` (ingest)
- `GET /v1/risks` (query-api)
- `GET /v1/alerts` (query-api)
- `PATCH /v1/alerts/{id}/ack` (query-api)
- `PATCH /v1/alerts/{id}/resolve` (query-api)
- `GET /v1/dashboard/summary` (query-api)

Detailed API examples: `docs/api.md`

## Production Launch Recommendation

For international scale, deploy on Kubernetes using managed data services:

- Kubernetes: GKE / EKS / AKS
- Kafka: Confluent Cloud or AWS MSK
- Postgres: Cloud SQL / RDS / AlloyDB
- Object storage for long-term event archive: S3 / GCS
- Observability: OpenTelemetry + Prometheus + Grafana + Loki

See `docs/deployment.md` and `docs/scaling.md`.

## License

MIT
