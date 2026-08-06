-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchants ADD COLUMN pic_name varchar(32);
ALTER TABLE merchants ADD COLUMN pic_job_title varchar(20);
ALTER TABLE merchants ADD COLUMN business_type varchar(20);
ALTER TABLE merchants ADD COLUMN business_structure varchar(20);
ALTER TABLE merchants ADD COLUMN business_country varchar(20);
ALTER TABLE merchants ADD COLUMN active bool default true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants DROP COLUMN pic_name;
ALTER TABLE merchants DROP COLUMN pic_job_title;
ALTER TABLE merchants DROP COLUMN business_type;
ALTER TABLE merchants DROP COLUMN business_structure;
ALTER TABLE merchants DROP COLUMN business_country;
ALTER TABLE merchants DROP COLUMN active;
-- +goose StatementEnd
