-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS inbound (
    id 				    VARCHAR(36) NOT NULL,
    reference_id 	    VARCHAR(36) AS (client->>'$.reference_id') STORED,
    origin_id 		    VARCHAR(36) AS (client->>'$.origin_id') STORED,
    trace_id 		    VARCHAR(36) AS (client->>'$.trace_id') STORED,
    ip                  VARCHAR(60) NOT NULL DEFAULT '',
    client 			    JSON NOT NULL,
    `method`		    VARCHAR(10) NOT NULL,
    url				    TEXT NOT NULL,
    headers			    JSON NOT NULL,
    body			    JSON NULL,
    status_code 	    SMALLINT UNSIGNED NULL,
    response_time_ms	FLOAT NOT NULL DEFAULT 0,
    response_body	    JSON NULL,
    metadata            JSON NULL,
    snap_compatibility  TINYINT DEFAULT 0,
    created_at		    timestamp NOT NULL,
    updated_at		    timestamp NOT NULL,
    PRIMARY KEY (id, created_at),
    KEY outbound_origin_id_idx (origin_id),
    KEY outbound_trace_id_idx (trace_id),
    KEY outbound_reference_id_created_at_comp_idx (reference_id, created_at)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inbound;
-- +goose StatementEnd

