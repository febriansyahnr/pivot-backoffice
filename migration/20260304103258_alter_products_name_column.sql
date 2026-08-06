-- +goose Up
-- +goose StatementBegin
ALTER TABLE products MODIFY COLUMN name varchar(60) NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products MODIFY COLUMN name varchar(20) NOT NULL;
-- +goose StatementEnd
