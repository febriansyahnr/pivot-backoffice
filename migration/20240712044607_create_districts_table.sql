-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `districts` (
	`id`			SMALLINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
	`city_id`	 	SMALLINT UNSIGNED NOT NULL,
	`name`			VARCHAR(30) NOT NULL,
	`created_at`	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	KEY `districts_city_id_idx` (`city_id`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `districts`;
-- +goose StatementEnd
