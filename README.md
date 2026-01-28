
# Orders
## Docker (Postgres + Kafka)
```bash
docker compose up -d
docker compose ps
```
After running docker, the kafka ui is located here:
- [Kafka UI](http://localhost:8080)
## Apply migrations:
```bash
psql postgres://app:app@localhost:5433/app -f migrations/001_init.up.sql
```
## Run API
```bash
go run ./cmd/api
```
After launching the API, the swagger ui is located here:
- [Swagger UI](http://localhost:8000/docs/)
## Run Worker
```bash
go run ./cmd/worker
```
