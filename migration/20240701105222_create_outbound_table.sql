-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS outbound (
	id 				VARCHAR(36) NOT NULL PRIMARY KEY,
	request_id		VARCHAR(36) AS (client->>'$.request_id') STORED,
	reference_id 	VARCHAR(36) AS (client->>'$.reference_id') STORED,
	origin_id 		VARCHAR(36) AS (client->>'$.origin_id') STORED,
	client 			JSON NOT NULL,
	`date`			timestamp NOT NULL,
	`method`		VARCHAR(10) NOT NULL,
	url				TEXT NOT NULL,
	headers			JSON NOT NULL,
	body			JSON NULL,
	status_code 	SMALLINT UNSIGNED NULL,
	response_time	VARCHAR(50) NULL,
	response_body	JSON NULL,
	error_message 	TEXT NULL,
	KEY outbound_origin_id_idx (origin_id),
	KEY outbound_request_id_idx (request_id),
	KEY outbound_reference_id_date_comp_idx (reference_id, `date`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbound;
-- +goose StatementEnd
