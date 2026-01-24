# Orders (Go + Postgres + Kafka)

Pet-project: API + worker.  
Пока что: базовый каркас + /health.

## Run API
```bash
go run ./cmd/api
```
## Infra (Postgres + Kafka)
```bash
docker compose up -d
docker compose ps
```