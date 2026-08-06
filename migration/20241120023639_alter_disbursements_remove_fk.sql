-- +goose Up
-- +goose StatementBegin
ALTER TABLE backend_portal.disbursements DROP FOREIGN KEY disbursements_ibfk_1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE backend_portal.disbursements
  ADD CONSTRAINT `disbursements_ibfk_1` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`uuid`);
-- +goose StatementEnd
