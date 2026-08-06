-- +goose Up
-- +goose StatementBegin
CREATE TABLE `callback_logs` (
  `uuid` char(36) NOT NULL,
  `callback_id` char(36) NOT NULL,
  `request` text NOT NULL,
  `response` text NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `callback_logs`;
-- +goose StatementEnd
