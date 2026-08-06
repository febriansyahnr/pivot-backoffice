-- +goose Up
-- +goose StatementBegin
INSERT INTO callback_masters (
	uuid, name, description, visibility, whitelisted_merchant_ids, created_at, updated_at, deleted_at
)
VALUES(
	'a53c8e50-53a3-4c3e-a9ee-c03684ca9880', 'SUB_ACCOUNT_REGISTRATION', 'Send callback whenever there is update in the sub-merchant status OR KYC status', 'PUBLIC', NULL, NOW(), NOW(), NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM callback_logs WHERE callback_id = 'a53c8e50-53a3-4c3e-a9ee-c03684ca9880';
DELETE FROM callbacks WHERE callback_master_id = 'a53c8e50-53a3-4c3e-a9ee-c03684ca9880';
DELETE FROM callback_masters WHERE uuid = 'a53c8e50-53a3-4c3e-a9ee-c03684ca9880';
-- +goose StatementEnd
