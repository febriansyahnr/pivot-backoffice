package schemas

func District() string {
	return `CREATE TABLE IF NOT EXISTS districts (
		id smallint unsigned NOT NULL AUTO_INCREMENT,
		city_id smallint unsigned NOT NULL,
		name varchar(30) NOT NULL,
		created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY districts_city_id_idx (city_id)
	) ENGINE=InnoDB;`
}
