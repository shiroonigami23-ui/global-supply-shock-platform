# Architecture

## Data Flow

1. Signal producers call `ingest` (`POST /v1/signals`) or simulation endpoint.
2. `ingest` publishes `SignalEvent` into Kafka topic `signals.raw`.
3. `risk-engine` consumes `signals.raw`, updates rolling window state, computes `RiskEvent`, persists it, and publishes to `risk.scored`.
4. `alert-service` consumes `risk.scored`, applies threshold/cooldown policy, persists alerts.
5. `query-api` serves risk and alert views to dashboards and external systems.

## Event Contracts

Defined in `internal/contracts/events.go`.

- `SignalEvent`
- `RiskEvent`
- `AlertRecord`

## Storage

PostgreSQL tables:

- `risk_events`
- `alerts`

Migration files are embedded from `internal/storage/sql` and executed on `query-api` startup.

## Fault Isolation

- Kafka decouples producers from processors.
- Each consumer service uses distinct consumer groups.
- Service restarts do not lose committed events.

## Extension Points

- Replace heuristic scoring with ML inference.
- Add tenant dimensions (`tenant_id`) to events and DB schemas.
- Add webhook/pager integration in alert-service.
