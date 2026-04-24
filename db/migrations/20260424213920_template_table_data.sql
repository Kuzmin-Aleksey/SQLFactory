-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
ALTER TABLE `SQLFactory`.`templates`
    ADD COLUMN `table_data` TEXT NOT NULL AFTER `query`;
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
