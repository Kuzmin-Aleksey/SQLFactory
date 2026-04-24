-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
ALTER TABLE `SQLFactory`.`history`
    ADD COLUMN `previous_id` INT NULL AFTER `user_id`;

CREATE TABLE `llm_contexts` (
    `id` int NOT NULL AUTO_INCREMENT,
    `history_id` int NOT NULL,
    `previous_id` int DEFAULT NULL,
    `role` varchar(45) NOT NULL,
    `content` text NOT NULL,
    PRIMARY KEY (`id`),
    KEY `context_to_history_idx` (`history_id`),
    CONSTRAINT `context_to_history` FOREIGN KEY (`history_id`) REFERENCES `history` (`id`) ON DELETE CASCADE
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
