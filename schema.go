package mysqlschema

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

func buildLatest(db *sql.DB, name, dir string, version int) (err error) {
	dir = path.Join(dir, "latest.sql")
	code, er := ioutil.ReadFile(dir)
	if er != nil {
		if os.IsNotExist(er) {
			return fmt.Errorf("latest.sql not found by table %s", name)
		}
		return er
	}
	if !strings.Contains(string(code), name) {
		return fmt.Errorf("table name %s not found in latest.sql", name)
	}
	log.Printf("upgrade latest schema of table %s...", name)
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
	_, er = tx.Exec(string(code))
	if er != nil {
		return er
	}
	_, er = tx.Exec("INSERT INTO __tb_schema__(name, schema_version) VALUES(?, ?)",
		name, version)
	if er != nil {
		return er
	}
	_, er = tx.Exec(`INSERT INTO __tb_schema_logs__(name, class, current_version, upgrade_version, code)
		VALUES(?, ?, ?, ?, ?)`, name, 0, 0, version, string(code))
	return er
}
