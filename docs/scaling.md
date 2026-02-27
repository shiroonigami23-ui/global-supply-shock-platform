# Scaling Notes

## Throughput

- Partition Kafka topics by `country|commodity` key.
- Scale `risk-engine` and `alert-service` by consumer group instance count.
- Use batch writes to Postgres for high event rates.

## Latency Targets

- Ingest API p95 < 50ms
- Risk scoring end-to-end < 2s
- Alert creation after threshold crossing < 5s

## Data Growth

- Partition `risk_events` by day/month.
- Archive old events to object storage.
- Keep `alerts` hot in primary DB and age out closed alerts by policy.

## Hardening Roadmap

1. Add idempotency keys and dedupe table.
2. Add schema registry and versioned contracts.
3. Add tenant-aware quotas and rate limits.
4. Add canary scoring model and drift monitoring.
