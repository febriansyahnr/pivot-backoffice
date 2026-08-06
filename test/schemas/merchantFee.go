package schemas

func MerchantFee() string {
	return `CREATE TABLE IF NOT EXISTS merchant_fees (
		uuid varchar(36) NOT NULL,
		merchant_id varchar(36) NOT NULL,
		reference varchar(100) NOT NULL,
		payment_method varchar(36) DEFAULT NULL,
		channel varchar(75) DEFAULT NULL,
		processor varchar(60) DEFAULT NULL,
		amount_type varchar(20) NOT NULL DEFAULT 'AMOUNT',
		amount decimal(18,2) NOT NULL DEFAULT '0.00',
		max_fee_amount decimal(18,2) DEFAULT NULL,
		percentage decimal(5,2) NOT NULL DEFAULT '0.00',
		reference_type varchar(30) NOT NULL DEFAULT '',
		deduction_type varchar(20) NOT NULL DEFAULT 'DIRECT',
		deduction_day tinyint unsigned DEFAULT NULL,
		deduction_last_date timestamp NULL DEFAULT NULL,
		tax_type varchar(20) NOT NULL DEFAULT '',
		tax_percentage decimal(5,2) NOT NULL DEFAULT '0.00',
		settlement_configs json DEFAULT NULL,
		tiering_model varchar(20) DEFAULT NULL,
		settlement_model varchar(20) DEFAULT NULL,
		settlement_method varchar(36) DEFAULT NULL,
		tiering_type varchar(20) DEFAULT NULL,
		tiering_configs json DEFAULT NULL,
		created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		deleted_at timestamp NULL DEFAULT NULL,
		PRIMARY KEY (uuid),
		KEY merchant_id_IDX (merchant_id) USING BTREE
	) ENGINE=InnoDB;`
}
