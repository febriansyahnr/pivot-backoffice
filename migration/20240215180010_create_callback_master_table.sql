-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `callback_masters` (
  `uuid` char(36) NOT NULL,
  `name` varchar(50) NOT NULL,
  `description` text NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `callback_masters`;
-- +goose StatementEnd
