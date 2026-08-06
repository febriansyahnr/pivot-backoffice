-- +goose Up
-- +goose StatementBegin
ALTER TABLE activity_logs
DROP PRIMARY KEY;

ALTER TABLE activity_logs
ADD PRIMARY KEY (id, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE activity_logs
DROP PRIMARY KEY;

ALTER TABLE activity_logs
ADD PRIMARY KEY (id);
-- +goose StatementEnd
