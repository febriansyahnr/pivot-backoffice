package schemas

func AccountTransaction() string {
	return `
		CREATE TABLE IF NOT EXISTS account_transactions (
		  uuid char(36) NOT NULL,
		  reference_id varchar(50) NOT NULL,
		  merchant_id char(36) NOT NULL,
		  balance_id char(36) NOT NULL,
		  currency varchar(3) NOT NULL,
		  credit decimal(19,4) NOT NULL DEFAULT '0',
		  debit decimal(19,4) NOT NULL DEFAULT '0',
		  type varchar(20) NOT NULL,
		  channel varchar(18) NOT NULL,
		  status varchar(7) NOT NULL,
		  reason_type varchar(100) NULL,
		  reason_description varchar(255) NULL,
		  remarks text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
		  transaction_timestamp timestamp NOT NULL,
		  settlement_at timestamp NULL,
		  settlement_status varchar(50) NULL,
		  additional_info JSON NULL,
		  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		  deleted_at timestamp NULL DEFAULT NULL,
		  PRIMARY KEY (uuid)
		);
	`
}
