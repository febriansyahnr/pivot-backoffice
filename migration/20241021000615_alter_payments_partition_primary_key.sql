-- +goose Up
-- +goose StatementBegin
ALTER TABLE payment_items 
DROP FOREIGN KEY `payment_items_ibfk_1`;

ALTER TABLE payments
DROP PRIMARY KEY;

ALTER TABLE payments
ADD PRIMARY KEY (uuid, created_at);

-- Add a unique index on uuid to satisfy foreign key requirements
ALTER TABLE payments
ADD UNIQUE INDEX idx_payments_uuid (uuid);

ALTER TABLE payment_items
ADD CONSTRAINT `payment_items_ibfk_1` FOREIGN KEY (`payment_id`) REFERENCES `payments` (`uuid`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payment_items 
DROP FOREIGN KEY `payment_items_ibfk_1`;

ALTER TABLE payments
DROP PRIMARY KEY;

-- Drop the unique index we added
ALTER TABLE payments
DROP INDEX idx_payments_uuid;

ALTER TABLE payments
ADD PRIMARY KEY (uuid);

ALTER TABLE payment_items
ADD CONSTRAINT `payment_items_ibfk_1` FOREIGN KEY (`payment_id`) REFERENCES `payments` (`uuid`);
-- +goose StatementEnd
