# API

Base URLs:

- ingest: `http://localhost:8081`
- query-api: `http://localhost:8080`

## POST /v1/signals

Publishes one raw signal.

Example request:

```json
{
  "source": "port_congestion",
  "country": "US",
  "region": "west",
  "commodity": "insulin",
  "metric_name": "queue_index",
  "metric_value": 84,
  "severity": 8,
  "confidence": 0.91
}
```

## POST /v1/simulate

Publishes N generated signals.

```json
{ "count": 200 }
```

## GET /v1/risks

Query params:

- `country` optional
- `commodity` optional
- `limit` optional (max 500)

## GET /v1/alerts

Query params:

- `status` optional (`open`, `acknowledged`, `resolved`)
- `limit` optional

## PATCH /v1/alerts/{id}/ack

Marks an alert as acknowledged.

## PATCH /v1/alerts/{id}/resolve

Marks an alert as resolved.

## GET /v1/dashboard/summary

Returns aggregate counts and average risk score for last 24h.
