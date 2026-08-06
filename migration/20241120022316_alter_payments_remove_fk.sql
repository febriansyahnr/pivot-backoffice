-- +goose Up
-- +goose StatementBegin
ALTER TABLE backend_portal.payments DROP FOREIGN KEY payments_ibfk_1;
ALTER TABLE backend_portal.payments DROP FOREIGN KEY payments_ibfk_3;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE backend_portal.payments
  ADD CONSTRAINT `payments_ibfk_1` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`uuid`),
  ADD CONSTRAINT `payments_ibfk_3` FOREIGN KEY (`payment_method_id`) REFERENCES `payment_methods` (`uuid`);
-- +goose StatementEnd
