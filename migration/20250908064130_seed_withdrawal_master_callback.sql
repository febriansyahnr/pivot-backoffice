-- +goose Up
-- +goose StatementBegin
INSERT INTO callback_masters (
	uuid, name, description, visibility, whitelisted_merchant_ids, created_at, updated_at, deleted_at
)
VALUES(
	'f0b9be63-8c7e-11f0-a6b3-42010a140011', 'WITHDRAWAL', 'Sends withdrawal transaction status notifications', 'PUBLIC', NULL, NOW(), NOW(), NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM callback_logs WHERE callback_id = 'f0b9be63-8c7e-11f0-a6b3-42010a140011';
DELETE FROM callbacks WHERE callback_master_id = 'f0b9be63-8c7e-11f0-a6b3-42010a140011';
DELETE FROM callback_masters WHERE uuid = 'f0b9be63-8c7e-11f0-a6b3-42010a140011';
-- +goose StatementEnd
