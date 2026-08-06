-- +goose Up
-- +goose StatementBegin
CREATE TABLE payment_captures (
      id                        VARCHAR(36)  NOT NULL PRIMARY KEY,
      payment_id                VARCHAR(36)  NOT NULL,   -- UUID in the payments table
      processor_reference_id    VARCHAR(60)  NULL, -- UUID in the processor transaction table
      status                    VARCHAR(10)  NOT NULL,   -- e.g. 'PENDING', 'SUCCESS', 'FAILED'
      release_remaining_amount  BOOLEAN      NOT NULL DEFAULT FALSE,
      currency                  VARCHAR(3)   NOT NULL,  -- e.g. 'IDR, USD, Or Etc',
      amount                    DECIMAL(65,2) NOT NULL,
      created_at                DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
      updated_at                DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
      INDEX idx_payment_captures_payment_id (payment_id),
      INDEX idx_payment_captures_processor_reference_id (processor_reference_id),
      INDEX idx_payment_captures_status_created_at (status, created_at)
) ENGINE=InnoDB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payment_captures;
-- +goose StatementEnd
