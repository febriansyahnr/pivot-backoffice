-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments ADD INDEX reason_type_invest_completed_comp_idx(reason_type, investigation_completed_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments DROP INDEX reason_type_invest_completed_comp_idx;
-- +goose StatementEnd
