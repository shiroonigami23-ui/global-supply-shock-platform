# INSTRUCTION

## 1. Prerequisites

- Docker Desktop
- Go 1.26+
- Git

## 2. Run End-to-End

```bash
docker compose up --build
```

This starts:

- Redpanda (Kafka API)
- PostgreSQL
- ingest
- risk-engine
- alert-service
- query-api

## 3. Seed Events

```bash
curl -X POST http://localhost:8081/v1/simulate -H "Content-Type: application/json" -d '{"count": 500}'
```

## 4. Read Results

```bash
curl "http://localhost:8080/v1/risks?limit=50"
curl "http://localhost:8080/v1/alerts?status=open&limit=50"
curl "http://localhost:8080/v1/dashboard/summary"
```

## 5. Alert Lifecycle

```bash
curl -X PATCH http://localhost:8080/v1/alerts/<alert-id>/ack
curl -X PATCH http://localhost:8080/v1/alerts/<alert-id>/resolve
```

## 6. Local Development in GoLand

1. Open folder `global-supply-shock-platform` in GoLand.
2. Let GoLand index modules.
3. Create run configs for:
   - `./cmd/query-api`
   - `./cmd/risk-engine`
   - `./cmd/alert-service`
   - `./cmd/ingest`
4. Set env vars from `.env.example`.

## 7. Build

```bash
go build ./...
```
