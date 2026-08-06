package schemas

func Province() string {
	return `CREATE TABLE IF NOT EXISTS provinces (
		id smallint unsigned NOT NULL AUTO_INCREMENT,
		name varchar(30) NOT NULL,
		created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id)
	) ENGINE=InnoDB;`
}
