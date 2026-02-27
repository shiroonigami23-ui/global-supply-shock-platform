# API

## Base URLs

- Local frontend proxy: `http://localhost:3000`
- Direct query-api: `http://localhost:8080`
- Direct ingest: `http://localhost:8081`

## Ingest

### POST /v1/signals

Publish one raw signal.

Example:

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

### POST /v1/simulate

Generate N random events for demo load.

```json
{ "count": 200 }
```

## Query API

### GET /v1/risks

Query params:

- `country` optional
- `commodity` optional
- `limit` optional (max 500)

### GET /v1/alerts

Query params:

- `status` optional (`open`, `acknowledged`, `resolved`)
- `limit` optional

### PATCH /v1/alerts/{id}/ack

Mark alert as acknowledged.

### PATCH /v1/alerts/{id}/resolve

Mark alert as resolved.

### GET /v1/dashboard/summary

Returns top-level KPI metrics.

### GET /v1/dashboard/timeseries?hours=24

Returns hourly trend points:

- `bucket_start`
- `avg_risk_score`
- `risk_events`
- `open_alerts_opened`

### GET /v1/dashboard/hotspots?hours=24&limit=20

Returns highest-risk geo+commodity groups with active alert count.
