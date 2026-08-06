-- +goose Up
-- +goose StatementBegin
ALTER TABLE customers MODIFY COLUMN uuid varchar(100) NOT NULL DEFAULT '',
 MODIFY COLUMN merchant_id varchar(100) NOT NULL DEFAULT '',
 MODIFY COLUMN email varchar(60) DEFAULT '',
 MODIFY COLUMN phone_number varchar(20) NOT NULL DEFAULT '',
 MODIFY COLUMN created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
 MODIFY COLUMN updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 MODIFY COLUMN last_name varchar(50) DEFAULT '',
 MODIFY COLUMN business_name varchar(255) DEFAULT '',
 MODIFY COLUMN city varchar(100) DEFAULT '',
 MODIFY COLUMN country varchar(2) DEFAULT '',
 MODIFY COLUMN address_line1 varchar(255) DEFAULT '',
 MODIFY COLUMN address_line2 varchar(255) DEFAULT '',
 MODIFY COLUMN postal_code varchar(10) DEFAULT '',
 MODIFY COLUMN state varchar(100) DEFAULT '',
-- ALTER TABLE customers MODIFY COLUMN metadata longtext DEFAULT '',
 MODIFY COLUMN first_name varchar(255) NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE customers 
 MODIFY uuid varchar(100) NOT NULL,
 MODIFY merchant_id varchar(100) NOT NULL,
 MODIFY email varchar(60) NULL DEFAULT NULL,
 MODIFY phone_number varchar(20) NULL DEFAULT NULL,
 MODIFY created_at datetime NOT NULL,
 MODIFY updated_at datetime NOT NULL,
 MODIFY last_name varchar(50) NULL DEFAULT NULL,
 MODIFY business_name varchar(255) NULL DEFAULT NULL,
 MODIFY city varchar(100) NULL DEFAULT NULL,
 MODIFY country varchar(2) NULL DEFAULT NULL,
 MODIFY address_line1 varchar(255) NULL DEFAULT NULL,
 MODIFY address_line2 varchar(255) NULL DEFAULT NULL,
 MODIFY postal_code varchar(10) NULL DEFAULT NULL,
 MODIFY state varchar(100) NULL DEFAULT NULL,
 MODIFY metadata longtext NULL,
 MODIFY first_name varchar(255) NULL DEFAULT NULL;
-- +goose StatementEnd
