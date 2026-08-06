-- +goose Up
-- +goose StatementBegin
ALTER TABLE menus ADD type varchar(12) DEFAULT '' NOT NULL AFTER name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE menus DROP COLUMN type;
-- +goose StatementEnd
