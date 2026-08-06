package schemas

func City() string {
	return `CREATE TABLE IF NOT EXISTS cities (
		id smallint unsigned NOT NULL AUTO_INCREMENT,
		province_id smallint unsigned NOT NULL,
		name varchar(30) NOT NULL,
		created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY cities_province_id_idx (province_id)
	) ENGINE=InnoDB;`
}
