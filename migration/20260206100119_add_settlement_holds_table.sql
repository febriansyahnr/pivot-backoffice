-- +goose Up
-- +goose StatementBegin
CREATE TABLE settlement_holds(
	uuid				VARCHAR(36) NOT NULL PRIMARY KEY,
	merchant_id			VARCHAR(36) NOT NULL,
    payment_id          VARCHAR(36) NOT NULL,
    status              VARCHAR(20) NOT NULL,
    created_by			CHAR(50) NOT NULL,
	created_at			TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at			TIMESTAMP NOT NULL,
	deleted_at			TIMESTAMP NULL,
	KEY payment_id_idx(payment_id)
);

CREATE TABLE settlement_hold_histories(
	uuid				VARCHAR(36) NOT NULL PRIMARY KEY,
	settlement_hold_id  VARCHAR(36) NOT NULL,
    status              VARCHAR(20) NOT NULL,
    reason              VARCHAR(100) NOT NULL,
    created_by			CHAR(50) NOT NULL,
	created_at			TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at			TIMESTAMP NULL,
	KEY settlement_hold_id_idx(settlement_hold_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE settlement_holds;
DROP TABLE settlement_hold_histories;
-- +goose StatementEnd
