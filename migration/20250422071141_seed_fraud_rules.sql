-- +goose Up
-- +goose StatementBegin
INSERT INTO backend_portal.fraud_rules
(uuid, rule_name, `condition`, priority, weight, is_active, reference_type, provider)
VALUES (
  'c714b0cc-8974-4cfe-805f-58f08d2e8ca8',
  'FRAUD_NET',
  'default_condition',
  1,
  1.0,
  1,
  '["ANY"]',
  'FRAUD_NET'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM backend_portal.fraud_rules
WHERE uuid = 'c714b0cc-8974-4cfe-805f-58f08d2e8ca8';
-- +goose StatementEnd
