-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_creditcards_reference_id ON creditcards(reference_id);
CREATE INDEX idx_creditcards_processor_reference_number ON creditcards(processor_reference_number);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_creditcards_reference_id ON creditcards;
DROP INDEX IF EXISTS idx_creditcards_processor_reference_number ON creditcards;
-- +goose StatementEnd
