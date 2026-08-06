-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments ADD payment_url varchar(255) DEFAULT '' NOT NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE payments CHANGE payment_url payment_url varchar(255) DEFAULT '' NOT NULL AFTER metadata;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE payments ADD expired_at datetime NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE payments CHANGE expired_at expired_at datetime NULL AFTER created_at;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE payments MODIFY COLUMN status varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'PENDING, SUCCESS, FAILED, ACTIVE, INACTIVE, WAITING_FOR_PAYMENT, EXPIRED';
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments DROP COLUMN payment_url;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE payments DROP expired_at;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE payments MODIFY COLUMN status varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'PENDING, SUCCESS, FAILED, ACTIVE, INACTIVE';
-- +goose StatementEnd

