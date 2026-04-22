set shell := ["pwsh.exe", "-CommandWithArgs"]
set dotenv-load := true

MIGRATION_DIR := "db/migrations"
TEST_DOCKER_COMPOSE := "docker compose -f tests/docker-compose.yml"

lint:
    golangci-lint run -D errcheck

goose-create NAME:
    goose -v -dir {{MIGRATION_DIR}} create {{NAME}} sql

goose-up:
    goose -v -dir {{MIGRATION_DIR}} mysql "{{env('MYSQL_USER', 'root')}}:{{env('MYSQL_PASSWORD')}}@/{{env('MYSQL_SCHEMA')}}?parseTime=true" up
    goose -v -dir {{MIGRATION_DIR}} mysql "{{env('MYSQL_USER', 'root')}}:{{env('MYSQL_PASSWORD')}}@/{{env('MYSQL_SCHEMA')}}?parseTime=true" status

goose-down:
    goose -v -dir {{MIGRATION_DIR}} mysql "{{env('MYSQL_USER', 'root')}}:{{env('MYSQL_PASSWORD')}}@/{{env('MYSQL_SCHEMA')}}?parseTime=true" down
    goose -v -dir {{MIGRATION_DIR}} mysql "{{env('MYSQL_USER', 'root')}}:{{env('MYSQL_PASSWORD')}}@/{{env('MYSQL_SCHEMA')}}?parseTime=true" status


install_deps:
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
    go install github.com/pressly/goose/v3/cmd/goose@latest