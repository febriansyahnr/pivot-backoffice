package schemas

func User() string {
	return `
		CREATE TABLE IF NOT EXISTS users (
		  uuid varchar(255) NOT NULL,
		  name varchar(255) NOT NULL,
		  email varchar(255) NOT NULL,
		  status varchar(20) DEFAULT 'ACTIVE',
		  password varchar(150) NOT NULL,
		  pin_hash varchar(255) DEFAULT NULL,
		  blocked_at datetime DEFAULT NULL,
		  merchant_id varchar(255) DEFAULT NULL,
		  refresh_token varchar(255) DEFAULT NULL,
		  is_change_password tinyint(1) NOT NULL DEFAULT '0',
		  created_at datetime NOT NULL,
		  updated_at datetime NOT NULL,
		  deleted_at datetime DEFAULT NULL,
		  last_login_at datetime DEFAULT NULL,
		  deactivate_at datetime DEFAULT NULL,
		  PRIMARY KEY (uuid),
		  UNIQUE KEY users_email_UNIQUE (email),
		  KEY users_merchant_id_IDX (merchant_id) USING BTREE
		) ENGINE=InnoDB
	`
}
