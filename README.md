# SQLFactory

## Запуск

1) Подготовить env для `docker compose`:

```bash
cp .env_example .env
```

2) Создать `config/config.yaml` (по примеру ниже) — это дефолтный путь, который ожидает `cmd/SQLFactory/main.go`.

Минимальный пример:

```yaml
http_server:
  addr: ":8080"

mysql:
  addr: "mysql:3306"
  user: "root"
  password: "pass"
  schema: "SQLFactory"

redis:
  addr: "redis:6379"
  password: "pass"

auth:
  sign_key: "change-me"

llm:
  name: "gigachat" # или "gemini"
  gigachat: { api_key: "" }
  gemini: { api_key: "" }
```

3) Запуск:

```bash
docker compose up --build
```

## Где смотреть эндпоинты

- **Swagger UI**: `http://localhost:8080/swagger/` (алиас: `/docs`)
- **OpenAPI YAML**: `http://localhost:8080/openapi.yaml`

