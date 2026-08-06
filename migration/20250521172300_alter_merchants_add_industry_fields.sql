-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchants 
    ADD COLUMN parent_industry VARCHAR(255) NULL,
    ADD COLUMN child_industry VARCHAR(255) NULL,
    ADD COLUMN mcc VARCHAR(255) NULL,
    ADD COLUMN country_of_entity VARCHAR(255) NULL,
    ADD COLUMN digital_status VARCHAR(255) NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants 
    DROP COLUMN parent_industry,
    DROP COLUMN child_industry,
    DROP COLUMN mcc,
    DROP COLUMN country_of_entity,
    DROP COLUMN digital_status;
-- +goose StatementEnd
