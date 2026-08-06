-- +goose Up
-- +goose StatementBegin
INSERT INTO payment_methods (uuid,`type`,sub_type,category,name,description,logo,acquirer,bank_name,activation_method,required_document,country_of_operation,supported_currency,processor,config,instructions,created_at,updated_at,deleted_at) VALUES
	 (uuid(),'VIRTUAL_TERMINAL','CARD','PAYMENT','Card','Card Payment','https://storage.googleapis.com/pg-staging-static-files/icon/partners/credit-cards.svg','harsya','','MANUAL',NULL,'ID','IDR','CREDIT_CARD_CORE_PROCESSOR','{"expiryConfig": {"unit": "MINUTES", "duration": 30}}',NULL,now(),now(),NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
