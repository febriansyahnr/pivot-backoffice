package schemas

func StatusHistory() string {
	return `CREATE TABLE IF NOT EXISTS status_histories (
		id varchar(36) NOT NULL,
		reference_type varchar(50) NOT NULL,
		reference_id varchar(64) NOT NULL,
		status varchar(50) NOT NULL,
		metadata json DEFAULT NULL,
		created_at datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		updated_at datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
		PRIMARY KEY (id),
		KEY idx_status_histories_ref (reference_type,reference_id)
	) ENGINE=InnoDB;`
}
