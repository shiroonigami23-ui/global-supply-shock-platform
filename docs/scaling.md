# Scaling

## Throughput Strategy

- Partition Kafka by `country|commodity`
- Scale consumers (`risk-engine`, `alert-service`) by partition count
- Keep stateless processors for easy horizontal scaling

## Latency Targets

- ingest API p95 < 75ms
- scoring pipeline < 2s end-to-end
- alert creation < 5s after threshold crossing

## Data Growth Management

- Partition `risk_events` by date in managed Postgres
- Archive old data to object storage
- Keep recent slices in primary DB for fast dashboard reads

## Backend Optimization Roadmap

1. Batch inserts in risk-engine
2. Introduce materialized views for hotspot queries
3. Add read cache for dashboard endpoints
4. Add dead-letter topic and replay tooling
5. Add tenant-aware quotas and auth
