-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `payment_method_merchant` (
    `uuid`			    VARCHAR(36) NOT NULL PRIMARY KEY,
    `merchant_id`	    VARCHAR(36) NOT NULL,
    `payment_method_id`	VARCHAR(36) NOT NULL,
    `is_active`			tinyint(1) NOT NULL DEFAULT '0',
    `created_at`	    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`	    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE `payment_method_merchant_merchant_id_payment_method_id_idx` (`merchant_id`, `payment_method_id`)
);

ALTER TABLE `payment_methods`
    ADD UNIQUE INDEX `payment_methods_category_type_acquirer_idx` (`category`,`type`,`acquirer`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `payment_methods` DROP INDEX `payment_methods_category_type_acquirer_idx`;

DROP TABLE IF EXISTS `payment_method_merchant`;
-- +goose StatementEnd
