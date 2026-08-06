-- +goose Up
-- +goose StatementBegin
INSERT INTO backend_portal.callback_masters
(uuid, name, description, visibility, whitelisted_merchant_ids, created_at, updated_at, deleted_at)
VALUES
('06813c42-4b04-40b1-ae0e-ff323c938149', 'WALLET_SNAP_QRIS_MPM', 'Wallet snap qris mpm callback url', 'PUBLIC', NULL, NOW(), NOW(), NULL),
('831dd2f1-49c3-4534-afbf-640c86449039', 'WALLET_SNAP_DIRECT_DEBIT', 'Wallet snap direct debit callback url', 'PUBLIC', NULL, NOW(), NOW(), NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM backend_portal.callback_masters WHERE name IN ('WALLET_SNAP_QRIS_MPM','WALLET_SNAP_DIRECT_DEBIT');
-- +goose StatementEnd
