
# Orders
## Swagger UI

1. Скачайте последнюю версию Swagger UI
```sh
wget https://github.com/swagger-api/swagger-ui/archive/refs/tags/v5.19.0.tar.gz
tar -xzf v5.19.0.tar.gz
```

2. Создайте директорию и скопируйте файлы
```sh
mkdir -p static/swagger-ui
cp -router swagger-ui-5.19.0/dist/* static/swagger-ui/
```

3. Удалите ненужные файлы
```sh
rm -rf swagger-ui-5.19.0
rm -f v5.19.0.tar.gz
```

4. Измените [swagger-initializer.js](./static/swagger-ui/swagger-initializer.js):
```
   window.ui = SwaggerUIBundle({
    url: "/openapi.yaml",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  });
```

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