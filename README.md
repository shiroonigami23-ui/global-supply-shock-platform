# Global Supply Shock Early-Warning Platform

A production-style, event-driven platform in Go that detects and alerts on disruption risk for essential commodities across countries.

## Included Stack

Backend services:

- `ingest` - receives raw disruption signals and publishes to Kafka
- `risk-engine` - computes rolling risk scores and stores events
- `alert-service` - opens alerts using threshold and cooldown rules
- `query-api` - serves dashboard/API reads and alert lifecycle updates

Frontend:

- `frontend` - responsive dashboard with summary cards, trend bars, hotspots, open alerts, and live risk feed

Data + stream:

- Redpanda (Kafka API)
- PostgreSQL

Delivery and operations:

- Dockerfiles and `docker-compose.yml`
- Kubernetes manifests (`deploy/k8s`)
- CI workflow for Go build
- Image publish workflow to GHCR

## Quick Start (Docker)

1. Start Docker Desktop.
2. Run:

```bash
docker compose up --build
```

3. Open:

- Dashboard: `http://localhost:3000`
- Query API: `http://localhost:8080`
- Ingest API: `http://localhost:8081`
- Redpanda Console: `http://localhost:8082`

4. Generate sample data:

```bash
curl -X POST http://localhost:8081/v1/simulate -H "Content-Type: application/json" -d '{"count": 500}'
```

## Main API Endpoints

- `POST /v1/signals` (ingest)
- `POST /v1/simulate` (ingest)
- `GET /v1/risks`
- `GET /v1/alerts`
- `PATCH /v1/alerts/{id}/ack`
- `PATCH /v1/alerts/{id}/resolve`
- `GET /v1/dashboard/summary`
- `GET /v1/dashboard/timeseries?hours=24`
- `GET /v1/dashboard/hotspots?hours=24&limit=20`

## Kubernetes

Full manifests are in `deploy/k8s/`:

- deployments: ingest, risk-engine, alert-service, query-api, dashboard
- services: ingest, query-api, dashboard
- autoscaling: query-api + risk-engine HPAs
- ingress: dashboard + API routes
- `secrets.example.yaml` template

Apply:

```bash
kubectl apply -k deploy/k8s
```

## Build and Test

```bash
go mod tidy
go build ./...
go test ./...
```

## Image Publishing

Workflow: `.github/workflows/docker-images.yml`

Builds and pushes:

- `ghcr.io/shiroonigami23-ui/global-supply-shock-platform/ingest`
- `ghcr.io/shiroonigami23-ui/global-supply-shock-platform/risk-engine`
- `ghcr.io/shiroonigami23-ui/global-supply-shock-platform/alert-service`
- `ghcr.io/shiroonigami23-ui/global-supply-shock-platform/query-api`
- `ghcr.io/shiroonigami23-ui/global-supply-shock-platform/dashboard`

## Recommended Production Launch

For international scale:

1. Kubernetes: EKS or GKE
2. Kafka: Confluent Cloud or AWS MSK
3. Postgres: RDS or Cloud SQL with read replicas
4. Edge: WAF + global load balancer + TLS
5. Observability: OpenTelemetry + Prometheus + Grafana + Loki
6. SLOs: risk processing latency, alert freshness, consumer lag, API p95

See `docs/deployment.md` and `docs/scaling.md`.

## License

MIT
