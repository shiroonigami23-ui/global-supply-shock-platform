# Deployment

## Recommended Production Topology

- Kubernetes cluster per major geography
- Managed Kafka (Confluent Cloud or MSK)
- Managed Postgres with read replicas
- Ingress controller + WAF + TLS cert manager
- Centralized logs, traces, and metrics

## Container Images

Push these images to GHCR or your registry:

- `ingest`
- `risk-engine`
- `alert-service`
- `query-api`
- `dashboard`

## Kubernetes Apply

```bash
kubectl apply -k deploy/k8s
```

## Required Secret Values

- `database-url`
- `kafka-brokers`

See `deploy/k8s/secrets.example.yaml`.

## Network Design

- Public routes:
  - dashboard host (UI + `/v1` API)
  - optional dedicated ingest host
- Internal-only services:
  - risk-engine
  - alert-service

## Operational Policies

- Pod autoscaling enabled for query-api and risk-engine
- Horizontal scaling through Kafka partitions and consumer groups
- Alerting on consumer lag, DB saturation, and API latency

## Disaster Recovery

- Multi-AZ Kafka and Postgres
- Snapshot + point-in-time restore for Postgres
- Replay from Kafka offsets for rebuilding downstream state
