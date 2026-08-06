-- +goose Up
-- +goose StatementBegin
CREATE TABLE `merchants` (
    `uuid` varchar(255) NOT NULL,
    `name` varchar(255) NOT NULL,
    `description` text,
    `logo` varchar(255) NOT NULL,
    `merchant_email` varchar(255) NOT NULL,
    `merchant_phone` varchar(255) NOT NULL,
    `pic_email` varchar(255) NOT NULL,
    `pic_phone` varchar(255) NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime,
    PRIMARY KEY (`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `merchants`;
-- +goose StatementEnd
