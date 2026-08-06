-- +goose Up
-- +goose StatementBegin
UPDATE merchants SET kyc_status='APPROVED' WHERE kyc_status IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE merchants SET kyc_status=NULL WHERE kyc_status = 'APPROVED';
-- +goose StatementEnd
