-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `provinces` (
	`id`			SMALLINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
	`name`			VARCHAR(30) NOT NULL,
	`created_at`	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `provinces`;
-- +goose StatementEnd
