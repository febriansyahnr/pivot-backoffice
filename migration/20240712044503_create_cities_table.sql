-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `cities` (
	`id`			SMALLINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
	`province_id` 	SMALLINT UNSIGNED NOT NULL,
	`name`			VARCHAR(30) NOT NULL,
	`created_at`	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	KEY `cities_province_id_idx` (`province_id`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `cities`;
-- +goose StatementEnd
