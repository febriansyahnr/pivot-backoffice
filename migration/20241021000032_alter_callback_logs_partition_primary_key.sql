-- +goose Up
-- +goose StatementBegin
ALTER TABLE callback_logs
DROP PRIMARY KEY;

ALTER TABLE callback_logs
ADD PRIMARY KEY (uuid, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE callback_logs
DROP PRIMARY KEY;

ALTER TABLE callback_logs
ADD PRIMARY KEY (uuid);
-- +goose StatementEnd
