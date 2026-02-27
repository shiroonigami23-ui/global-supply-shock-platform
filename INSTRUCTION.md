# INSTRUCTION

## Tooling Installed

Global tooling expected on this machine:

- Docker Desktop
- Go 1.26+
- kubectl
- Helm

If commands are not visible in an already-open terminal, restart the terminal session.

## Local End-to-End Run

```bash
docker compose up --build
```

Services:

- frontend dashboard: `http://localhost:3000`
- query-api: `http://localhost:8080`
- ingest: `http://localhost:8081`
- redpanda console: `http://localhost:8082`

## Seed Demo Data

```bash
curl -X POST http://localhost:8081/v1/simulate -H "Content-Type: application/json" -d '{"count": 500}'
```

## Kubernetes Deploy

1. Edit `deploy/k8s/secrets.example.yaml` with real broker/database values.
2. Apply manifests:

```bash
kubectl apply -k deploy/k8s
```

3. Validate:

```bash
kubectl get pods -n supply-shock
kubectl get svc -n supply-shock
kubectl get ingress -n supply-shock
```

## API Smoke Checks

```bash
curl http://localhost:8080/healthz
curl "http://localhost:8080/v1/dashboard/summary"
curl "http://localhost:8080/v1/dashboard/timeseries?hours=24"
curl "http://localhost:8080/v1/dashboard/hotspots?hours=24&limit=10"
```

## Development in GoLand

1. Open this folder in GoLand.
2. Configure run targets:
   - `./cmd/ingest`
   - `./cmd/risk-engine`
   - `./cmd/alert-service`
   - `./cmd/query-api`
3. Use `.env.example` for environment references.
