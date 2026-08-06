-- +goose Up
-- +goose StatementBegin
ALTER TABLE disbursements ADD COLUMN reference_id varchar(100) NOT NULL AFTER uuid;
ALTER TABLE disbursements ADD COLUMN remark varchar(60) NULL AFTER reason_description;
ALTER TABLE disbursements ADD CONSTRAINT disbursement_unique_reference_per_merchant UNIQUE (reference_id, merchant_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE disbursements DROP CONSTRAINT disbursement_unique_reference_per_merchant;
ALTER TABLE disbursements DROP COLUMN remark;
ALTER TABLE disbursements DROP COLUMN reference_id;
-- +goose StatementEnd
