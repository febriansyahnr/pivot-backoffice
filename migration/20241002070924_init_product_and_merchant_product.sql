-- +goose Up
-- +goose StatementBegin
CREATE TABLE `products` (
  `uuid` varchar(36) NOT NULL,
  `name` varchar(20) NOT NULL,
  `active` SMALLINT NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`uuid`)
);

-- Seed available products
INSERT INTO `products` (`uuid`, `name`, `active`, `created_at`, `updated_at`) 
VALUES
    (UUID(),'PAYMENT', 1, NOW(), NOW()),
    (UUID(),'DISBURSEMENT', 1, NOW(), NOW()),
    (UUID(),'PLATFORM', 1, NOW(), NOW()),
    (UUID(),'WALLET WHITELABEL', 1, NOW(), NOW());


CREATE TABLE `merchant_selected_products` (
  `uuid` varchar(36) NOT NULL,
  `merchant_id` varchar(36) NOT NULL,
  `product_id` varchar(36) NOT NULL,
  `active` SMALLINT NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `merchant_products_id_composite_IDX` (`merchant_id`, `product_id`) USING BTREE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE `merchant_selected_products`;
DROP TABLE `products`;

-- +goose StatementEnd
