-- +goose Up
-- +goose StatementBegin
CREATE TABLE installment_plans (
    uuid varchar(36),
    merchant_id varchar(36),
    acquirer varchar(30),
    settlement_type varchar(30),
    payment_method varchar(20),
    title varchar(100),
    description varchar(255),
    tenor tinyint,
    status varchar(20),
    metadata json,
    created_at DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME  NULL,
    PRIMARY KEY (`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE installment_plans;
-- +goose StatementEnd
