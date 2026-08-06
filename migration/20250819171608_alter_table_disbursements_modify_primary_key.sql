-- +goose Up
-- +goose StatementBegin
ALTER TABLE `disbursements` DROP PRIMARY KEY, ADD PRIMARY KEY (`uuid`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `disbursements` DROP PRIMARY KEY, ADD PRIMARY KEY (`uuid`, `created_at`);
-- +goose StatementEnd
