-- +goose Up
-- +goose StatementBegin
ALTER TABLE creditcards ADD merchant_id varchar(36) DEFAULT '' NOT NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE creditcards CHANGE merchant_id merchant_id varchar(36) DEFAULT '' NOT NULL AFTER card_data;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE creditcards DROP COLUMN merchant_id;
-- +goose StatementEnd
