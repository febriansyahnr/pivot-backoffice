-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts 
MODIFY last_update_balance_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
ALGORITHM=COPY,
LOCK=SHARED;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts 
MODIFY last_update_balance_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
ALGORITHM=COPY,
LOCK=SHARED;
-- +goose StatementEnd
