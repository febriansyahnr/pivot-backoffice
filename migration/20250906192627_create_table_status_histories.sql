-- +goose Up
-- +goose StatementBegin
CREATE TABLE status_histories (
    id              VARCHAR(36)  NOT NULL PRIMARY KEY,
    reference_type  VARCHAR(50)  NOT NULL,   -- e.g. 'payment', 'withdrawal', 'disbursement'
    reference_id    VARCHAR(64)  NOT NULL,   -- ID in the referenced table
    status          VARCHAR(50)  NOT NULL,   -- e.g. 'CREATED', 'PROCESSING', 'PAID', 'FAILED'
    metadata        JSON         NULL,       -- extra data like requestId, source, actor, etc.
    created_at      DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at      DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    INDEX idx_status_histories_ref   (reference_type, reference_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE status_histories;
-- +goose StatementEnd
