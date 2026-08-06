-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `daily_account_transactions`(
    `id`				VARCHAR(36) NOT NULL PRIMARY KEY,
    `account_id`        VARCHAR(36) NOT NULL,
    `date`              DATE NOT NULL,
    `timezone`          VARCHAR(50) NOT NULL,
    `beg_balance`       DECIMAL(19,4) NOT NULL,
    `debit_trx`         INT NOT NULL,
    `debit_amount`      DECIMAL(19,4) NOT NULL,
    `credit_trx`        INT NOT NULL,
    `credit_amount`     DECIMAL(19,4) NOT NULL,
    `eod_balance`       DECIMAL(19,4) NOT NULL,
    `created_at`		TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `updated_at`		TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    KEY `daily_account_transactions_account_id_idx`(`account_id`),
    UNIQUE KEY `unique_account_date_timezone` (`account_id`, `date`, `timezone`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `daily_account_transactions`;
-- +goose StatementEnd