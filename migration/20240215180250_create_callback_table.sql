-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `callbacks` (
  `uuid` char(36) NOT NULL,
  `callback_master_id` char(36) NOT NULL,
  `merchant_id` char(36) NOT NULL,
  `url` text NOT NULL,
  `description` text NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `callbacks`;
-- +goose StatementEnd
