package schemas

func BulkDisbursement() string {
	return `
	CREATE TABLE bulk_disbursements (
		uuid varchar(100) NOT NULL,
		merchant_id varchar(100) NOT NULL,
		file varchar(255) NOT NULL,
		file_failed varchar(255) DEFAULT NULL,
		file_rejected varchar(255) DEFAULT NULL,
		status varchar(20) NOT NULL COMMENT 'WAITING, IN_PROGRESS, DONE, PENDING',
		created_by varchar(100) DEFAULT NULL,
		created_at datetime NOT NULL,
		updated_at datetime NOT NULL,
		deleted_at datetime DEFAULT NULL,
		PRIMARY KEY (uuid),
		KEY bulk_disbursements_merchant_id_status_created_at_IDX (merchant_id,status,created_at) USING BTREE
	) ENGINE=InnoDB`
}
