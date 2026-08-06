-- +goose Up
-- +goose StatementBegin
alter table account_transactions modify column updated_at timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6);
 
alter table account_transactions modify column created_at timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table account_transactions modify column updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP();
 
alter table account_transactions modify column created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP();

-- +goose StatementEnd
