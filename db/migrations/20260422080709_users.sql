-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
create table users
(
    id       int auto_increment
        primary key,
    name     varchar(255) not null,
    email    varchar(255) not null,
    password varchar(512) not null
);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
