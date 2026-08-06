-- +goose Up
-- +goose StatementBegin
UPDATE backend_portal.industries i 
	SET i.common_mcc = "7399", i.updated_at = NOW()
WHERE i.parent_industry = "Financial services" AND i.child_industry IN ("Other financial services", "Aggregator/Payment Reseller");
UPDATE backend_portal.industries i 
	SET i.common_mcc = "5399", i.updated_at = NOW()
WHERE i.child_industry IN ("Department Stores", "Grocery Stores", "Miscellaneous and Specialty Retail", "Book Stores", "Office Supplies", "Furniture, home decor & home appliances", "Alcohol","Other retail");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE industries SET parent_industry='Financial services', child_industry='Other financial services', risk_level='High', mcc='6051', common_mcc='6051', created_at='2025-06-16 10:08:09.417507', updated_at='2025-06-16 10:08:09.417507' WHERE uuid='cdf35910-4a99-11f0-91a6-42010a9c01c0';
UPDATE industries SET parent_industry='Financial services', child_industry='Aggregator/Payment Reseller', risk_level='High', mcc='6051', common_mcc='6051', created_at='2025-06-16 10:08:09.417507', updated_at='2025-06-16 10:08:09.417507' WHERE uuid='cdf3595c-4a99-11f0-91a6-42010a9c01c0';
UPDATE industries SET parent_industry='Retail', child_industry='Department Stores', risk_level='Low', mcc='5311', common_mcc='5999', created_at='2025-06-16 10:08:09.417507', updated_at='2025-06-16 10:08:09.417507' WHERE uuid='cdf36aa3-4a99-11f0-91a6-42010a9c01c0';
UPDATE industries SET parent_industry='Retail', child_industry='Grocery Stores', risk_level='Low', mcc='5411', common_mcc='5999', created_at='2025-06-16 10:08:09.417507', updated_at='2025-06-16 10:08:09.417507' WHERE uuid='cdf36b00-4a99-11f0-91a6-42010a9c01c0';
UPDATE industries SET parent_industry='Retail', child_industry='Miscellaneous and Specialty Retail', risk_level='Low', mcc='5999', common_mcc='5999', created_at='2025-06-16 10:08:09.417507', updated_at='2025-06-16 10:08:09.417507' WHERE uuid='cdf36bf8-4a99-11f0-91a6-42010a9c01c0';
UPDATE industries SET parent_industry='Retail', child_industry='Book Stores', risk_level='Low', mcc='5942', common_mcc='5999', created_at='2025-06-16 10:08:09.417507', updated_at='2025-06-16 10:08:09.417507' WHERE uuid='cdf36c4d-4a99-11f0-91a6-42010a9c01c0';
UPDATE industries SET parent_industry='Retail', child_industry='Office Supplies', risk_level='Low', mcc='5943', common_mcc='5999', created_at='2025-06-16 10:08:09.417507', updated_at='2025-06-16 10:08:09.417507' WHERE uuid='cdf36ca0-4a99-11f0-91a6-42010a9c01c0';
UPDATE industries SET parent_industry='Retail', child_industry='Furniture, home decor & home appliances', risk_level='Low', mcc='5712', common_mcc='5999', created_at='2025-06-16 10:08:09.417507', updated_at='2025-06-16 10:08:09.417507' WHERE uuid='cdf36cf8-4a99-11f0-91a6-42010a9c01c0';
UPDATE industries SET parent_industry='Retail', child_industry='Alcohol', risk_level='High', mcc='5921', common_mcc='5999', created_at='2025-06-16 10:08:09.417507', updated_at='2025-06-16 10:08:09.417507' WHERE uuid='cdf36d46-4a99-11f0-91a6-42010a9c01c0';
UPDATE industries SET parent_industry='Retail', child_industry='Other retail', risk_level='Medium', mcc='5999', common_mcc='5999', created_at='2025-06-16 10:08:09.417507', updated_at='2025-06-16 10:08:09.417507' WHERE uuid='cdf36da8-4a99-11f0-91a6-42010a9c01c0';
-- +goose StatementEnd
