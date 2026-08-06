-- +goose Up
-- +goose StatementBegin
-- Insert product
INSERT INTO `products` (`uuid`, `name`, `active`, `created_at`, `updated_at`)
VALUES (uuid(), 'CARD-FUNDED PAYOUT', 1, NOW(), NOW());

-- Update menu using the same UUID
UPDATE `menus`
SET `allowed_products` = JSON_ARRAY('CARD-FUNDED PAYOUT')
WHERE `slug` = 'card-funded-payout';

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DELETE FROM `products`
WHERE `name` = 'CARD-FUNDED PAYOUT';

UPDATE `menus`
SET `allowed_products` = NULL
WHERE `slug` = 'card-funded-payout';

-- +goose StatementEnd