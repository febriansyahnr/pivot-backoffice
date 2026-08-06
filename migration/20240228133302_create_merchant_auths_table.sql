-- +goose Up
-- +goose StatementBegin
CREATE TABLE `merchant_auths` (
     `uuid` varchar(100) NOT NULL,
     `merchant_id` varchar(100) NOT NULL,
     `secret` varchar(255) NOT NULL,
     `created_at` datetime NOT NULL,
     `updated_at` datetime NOT NULL,
     `deleted_at` datetime NULL,
     PRIMARY KEY (`uuid`),
     KEY `merchant_auths_merchant_id_IDX` (`merchant_id`) USING BTREE,
     FOREIGN KEY (`merchant_id`) REFERENCES `merchants`(`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `merchant_auths`;
-- +goose StatementEnd
