-- +goose Up
-- +goose StatementBegin
ALTER TABLE callback_logs ADD COLUMN event varchar(60) NULL AFTER callback_id;
ALTER TABLE callback_logs ADD COLUMN status varchar(20) NOT NULL DEFAULT "PENDING" AFTER response;
ALTER TABLE callback_logs ADD COLUMN retry smallint NOT NULL DEFAULT 0 AFTER status;
ALTER TABLE callback_logs ADD COLUMN updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER created_at;
ALTER TABLE callback_logs CHANGE COLUMN response response TEXT NULL;

ALTER TABLE callback_logs ADD INDEX `callback_logs_updated_at_IDX`(`updated_at`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE callback_logs DROP INDEX `callback_logs_updated_at_IDX`;

ALTER TABLE callback_logs DROP COLUMN event;
ALTER TABLE callback_logs DROP COLUMN status;
ALTER TABLE callback_logs DROP COLUMN retry;
ALTER TABLE callback_logs DROP COLUMN updated_at;
ALTER TABLE callback_logs CHANGE COLUMN response response TEXT NOT NULL;
-- +goose StatementEnd
