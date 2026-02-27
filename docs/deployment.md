# Deployment

## Recommended Production Stack

- Compute: Kubernetes (GKE/EKS/AKS)
- Event stream: Confluent Cloud Kafka or AWS MSK
- Database: managed PostgreSQL (Cloud SQL / RDS / AlloyDB)
- Secrets: cloud secret manager
- API edge: cloud load balancer + WAF
- Observability: OpenTelemetry + Prometheus + Grafana + Loki

## Multi-Region Pattern

- Active-active ingest endpoints per region.
- Kafka cluster with cross-region replication.
- Regional risk/alert workers reading local replicas.
- Query API with geo-routing and read replicas.

## Reliability Controls

- Consumer lag alerts
- Dead-letter topic for malformed events
- Backpressure with producer retry + circuit breaker
- Blue/green deployments for service updates

## Security Controls

- mTLS between services
- JWT/OAuth2 at API edge
- Row-level tenant segregation in DB
- Audit log for alert status changes
