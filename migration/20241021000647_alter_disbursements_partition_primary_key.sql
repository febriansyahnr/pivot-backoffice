-- +goose Up
-- +goose StatementBegin
ALTER TABLE disbursements
DROP PRIMARY KEY;

ALTER TABLE disbursements
ADD PRIMARY KEY (uuid, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE disbursements
DROP PRIMARY KEY;

ALTER TABLE disbursements
ADD PRIMARY KEY (uuid);
-- +goose StatementEnd
