# Architecture

## Services

- `ingest` (Go): accepts raw disruption signals and publishes to Kafka
- `risk-engine` (Go): consumes raw signals, computes risk scores, stores risk events, emits scored events
- `alert-service` (Go): consumes scored events, opens alerts based on threshold/cooldown policy
- `query-api` (Go): API for dashboard and operators (read risks/alerts, update alert state)
- `dashboard` (Nginx + static JS): frontend UI with live operational views

## Event Flow

1. Producers call `ingest /v1/signals`.
2. Ingest writes `SignalEvent` to `signals.raw`.
3. Risk-engine reads `signals.raw`, computes `RiskEvent`, stores to Postgres, writes `risk.scored`.
4. Alert-service reads `risk.scored`, creates `alerts` rows when score crosses threshold.
5. Query-api exposes read models and mutation endpoints.
6. Frontend reads query-api endpoints and allows alert acknowledgements/resolution.

## Storage

- Postgres tables:
  - `risk_events`
  - `alerts`

## Deployment Modes

- Local: Docker Compose
- Production: Kubernetes + managed Kafka + managed Postgres

## Reliability

- Kafka decouples ingest from processing and alerting
- Consumer groups allow horizontal scaling
- Stateless compute services support rolling deployment and failover
