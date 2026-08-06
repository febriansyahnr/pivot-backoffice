-- +goose Up
-- +goose StatementBegin
INSERT INTO `callback_masters` (`uuid`, `name`, `description`, `visibility`, `whitelisted_merchant_ids`, `created_at`, `updated_at`)
    VALUES (uuid(), 'MERCHANT_TOP_UP', 'Merchant Top Up', 'RESTRICTED', '[\"\"]', NOW(), NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM `callback_masters` WHERE `name` = 'MERCHANT_TOP_UP';
-- +goose StatementEnd
