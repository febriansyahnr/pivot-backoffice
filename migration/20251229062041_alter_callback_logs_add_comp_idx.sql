-- +goose Up
-- +goose StatementBegin
ALTER TABLE callback_logs 
	ADD INDEX callback_logs_callback_id_updated_comp_idx (callback_id, updated_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE callback_logs 
	DROP INDEX callback_logs_callback_id_updated_comp_idx;
-- +goose StatementEnd
