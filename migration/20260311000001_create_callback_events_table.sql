-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `callback_events` (
  `uuid` char(36) NOT NULL,
  `event` varchar(100) NOT NULL,
  `label` varchar(150) NOT NULL,
  `event_group` varchar(50) NOT NULL,
  `is_active` tinyint NOT NULL DEFAULT 1,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `callback_events_event_IDX` (`event`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `callback_events`;
-- +goose StatementEnd
