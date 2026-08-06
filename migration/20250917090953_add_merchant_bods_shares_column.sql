-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchant_bods ADD COLUMN shares DECIMAL(7,4) AFTER status;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchant_bods DROP COLUMN shares;
-- +goose StatementEnd
