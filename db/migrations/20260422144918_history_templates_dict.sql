-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE `SQLFactory`.`history` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `user_id` INT NOT NULL,
    `db` VARCHAR(512) NOT NULL,
    `create_at` TIMESTAMP NOT NULL,
    `title` VARCHAR(512) NOT NULL,
    `prompt` TEXT NOT NULL,
    `query` TEXT NOT NULL,
    `table_data` TEXT NOT NULL,
    `chart_type` TEXT NOT NULL,
    `reasoning` TEXT NOT NULL,
PRIMARY KEY (`id`),
INDEX `history_to_user_idx` (`user_id` ASC) VISIBLE,
CONSTRAINT `history_to_user`
    FOREIGN KEY (`user_id`)
        REFERENCES `SQLFactory`.`users` (`id`)
        ON DELETE CASCADE
        ON UPDATE NO ACTION);


CREATE TABLE `SQLFactory`.`templates` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `db` VARCHAR(512) NOT NULL,
    `title` VARCHAR(512) NOT NULL,
    `query` TEXT NOT NULL,
    `chart_type` TEXT NOT NULL,
PRIMARY KEY (`id`));

CREATE TABLE `dict` (
    `id` int NOT NULL AUTO_INCREMENT,
    `db` varchar(512) NOT NULL,
    `word` varchar(50) NOT NULL,
    `meaning` text NOT NULL,
PRIMARY KEY (`id`));

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
