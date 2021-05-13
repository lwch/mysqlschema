package mysqlschema

import "database/sql"

func prepare(db *sql.DB) (err error) {
	tx, er := db.Begin()
	if er != nil {
		return er
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
		}
	}()
	_, er = tx.Exec(`
	CREATE TABLE IF NOT EXISTS __tb_schema__(
		id int NOT NULL AUTO_INCREMENT,
		created datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		name varchar(256) NOT NULL DEFAULT '' COMMENT 'table name',
		schema_version int NOT NULL DEFAULT 0 COMMENT 'schema version',
		data_version int NOT NULL DEFAULT 0 COMMENT 'data version',
		PRIMARY KEY(id),
		UNIQUE KEY idx_uniq(name, schema_version, data_version)
	)Engine=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT 'current table schema'`)
	if er != nil {
		return err
	}
	_, er = tx.Exec(`
	CREATE TABLE IF NOT EXISTS __tb_schema_logs__(
		id int NOT NULL AUTO_INCREMENT,
		created datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
		name varchar(256) NOT NULL DEFAULT '' COMMENT 'table name',
		class tinyint NOT NULL DEFAULT 0 COMMENT '0=schema;1=data',
		current_version int NOT NULL DEFAULT 0 COMMENT 'current version',
		upgrade_version int NOT NULL DEFAULT 0 COMMENT 'upgrade version',
		code TEXT NOT NULL COMMENT 'upgrade code',
		PRIMARY KEY(id),
		INDEX idx_name_class(name, class)
	)Engine=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT 'table schema version change log'`)
	return er
}
