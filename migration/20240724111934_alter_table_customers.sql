-- +goose Up
-- +goose StatementBegin
ALTER TABLE `customers` 
    RENAME COLUMN full_name TO first_name,
    ADD COLUMN last_name VARCHAR(50) NULL,
    ADD COLUMN business_name VARCHAR(255) NULL,

    ADD COLUMN city VARCHAR(100) NULL,
    ADD COLUMN country VARCHAR(2) NULL,
    ADD COLUMN address_line1 VARCHAR(255) NULL,
    ADD COLUMN address_line2 VARCHAR(255) NULL,
    ADD COLUMN postal_code VARCHAR(10) NULL,
    ADD COLUMN state VARCHAR(100) NULL,

    ADD COLUMN metadata JSON NULL
;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE `customers` 
    RENAME COLUMN first_name TO full_name,
    DROP COLUMN last_name,
    DROP COLUMN business_name,

    DROP COLUMN city,
    DROP COLUMN country,
    DROP COLUMN address_line1,
    DROP COLUMN address_line2,
    DROP COLUMN postal_code,
    DROP COLUMN state,

    DROP COLUMN metadata,
    ADD COLUMN full_name VARCHAR(255) NOT NULL
;
-- +goose StatementEnd
