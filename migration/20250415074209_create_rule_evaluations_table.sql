-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `rule_evaluations` (
    `uuid` VARCHAR(36)  NOT NULL,
	`reference_id` CHAR(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
	`rule_id` VARCHAR(36) NOT NULL,
	`result` VARCHAR(100) NOT NULL,
	`score`  DOUBLE DEFAULT 0.0 NOT NULL,
	`reason` VARCHAR(255) NULL DEFAULT NULL,
	`evaluated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (`uuid`),
    KEY `rule_evaluations_reference_id_IDX` (`reference_id`) USING BTREE,
    KEY `rule_evaluations_rule_id_IDX` (`rule_id`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `rule_evaluations`;
-- +goose StatementEnd
