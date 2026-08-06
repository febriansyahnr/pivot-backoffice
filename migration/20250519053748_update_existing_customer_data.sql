-- +goose Up
-- +goose StatementBegin
UPDATE customers c 
SET phone_country_code = "+62"
WHERE c.phone_number != "-" AND (c.phone_country_code is null or c.phone_country_code = "");

UPDATE customers
SET phone_number = 
  CASE
    WHEN phone_number LIKE '0%' THEN SUBSTRING(phone_number FROM 2)  -- Remove '0'
    WHEN phone_number LIKE '062%' THEN SUBSTRING(phone_number FROM 4) -- Remove '062'
    WHEN phone_number LIKE '62%' THEN SUBSTRING(phone_number FROM 3)  -- Remove '62'
    ELSE phone_number
  END
WHERE phone_number LIKE '0%' 
   OR phone_number LIKE '62%' 
   OR phone_number LIKE '062%';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
