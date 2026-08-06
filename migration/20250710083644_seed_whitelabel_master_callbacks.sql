-- +goose Up
-- +goose StatementBegin
INSERT INTO backend_portal.callback_masters
(uuid, name, description, visibility, whitelisted_merchant_ids, created_at, updated_at, deleted_at)
VALUES(UUID(), 'WALLET', 'Wallet', 'PUBLIC', NULL, NOW(), NOW(), NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM backend_portal.callback_masters WHERE name IN ('WALLET');
-- +goose StatementEnd
