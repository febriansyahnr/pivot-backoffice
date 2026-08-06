-- +goose Up
-- +goose StatementBegin
ALTER TABLE backend_portal.payment_items DROP FOREIGN KEY payment_items_ibfk_1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE backend_portal.payment_items
  ADD CONSTRAINT `payment_items_ibfk_1` FOREIGN KEY (`payment_id`) REFERENCES `payments` (`uuid`);
-- +goose StatementEnd
