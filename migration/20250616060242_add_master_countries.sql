-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS countries (
    code                 VARCHAR(5) NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    name_id              VARCHAR(255) NOT NULL,
    created_at           timestamp NOT NULL,
    updated_at           timestamp NOT NULL,
    deleted_at           timestamp NULL,

    PRIMARY KEY (code)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS countries;
-- +goose StatementEnd
