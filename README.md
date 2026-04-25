# SQLFactory

## Запуск в тестовом режиме

1. Переименуйте .env_example в .env
    ```bash
    cp .env_example .env
    ```

2. Создать `config/config.yaml`:

    ```yaml
    http_server:
      addr: "0.0.0.0:8080"
      handle_timeout: 50s # запросы к llm выполняются долго
      disable_auth: true # отключаем авторизацию
      write_timeout: 0s
    
      log:
        max_request_content_len: 1024
        max_response_content_len: 512
        request_logging_content:
          - "application/json"
        response_logging_content:
          - "application/json"
    
    auth:
      access_token_ttl: 720h
      refresh_token_ttl: 720h
      access_token_sign_key: "qqq"
    
    log:
      path: "logs"
      level: "debug"
    
    redis:
      addr: "127.0.0.1:6379"
      password: "pass"
    
    mysql:
      addr: "127.0.0.1:3306"
      user: root
      password: pass
      schema: "SQLFactory"
      connect_timeout_sec: 10
    
    sql_runner:
      max_rows: 1000
    
    llm:
      name: "gemini" # название llm
    
      # данные для подключения:
      gemini:
        api_key: ""
    
      gigachat:
        api_key: ""
        model: "GigaChat-2-Pro"
    
    debug_user: true # создаем тестового пользователя
    ```

3. Добавьте файл orders.csv в /db/testdata (файл train.csv)

4. Запуск в docker
    ```bash
    docker compose up --build
    ```

## Данные для подключения к тестовой БД:

- `db_type`: postgres
- `host`:   test_db
- `port`: 5432
- `database`: test
- `user`: test
- `password`: pass

## API

- **Swagger UI**: `http://localhost:8080/swagger/`
- **OpenAPI YAML**: `http://localhost:8080/openapi.yaml`

