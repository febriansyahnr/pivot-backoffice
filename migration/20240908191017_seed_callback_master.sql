-- +goose Up
-- +goose StatementBegin
INSERT INTO `callback_masters` (`uuid`, `name`, `description`, `created_at`, `updated_at`) VALUES (uuid(), 'INTERNATIONAL_PAYOUT', 'International Payout', now() , now());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
