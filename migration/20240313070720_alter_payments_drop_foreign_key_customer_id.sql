-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments DROP FOREIGN KEY payments_ibfk_2;
ALTER TABLE payments DROP INDEX payments_customer_id_IDX;
ALTER TABLE payments MODIFY customer_id VARCHAR(100) NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments ADD CONSTRAINT payments_customer_id_IDX FOREIGN KEY (customer_id) REFERENCES customers(uuid);
ALTER TABLE payments MODIFY customer_id VARCHAR(100) NOT NULL;
-- +goose StatementEnd