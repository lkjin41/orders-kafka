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
## Database schema

Tables:
- orders
- outbox (transactional outbox for Kafka)
- processed_events (idempotency)
- order_activity (read model)

Apply migrations:

```bash
psql postgres://app:app@localhost:5432/app -f migrations/001_init.up.sql
```