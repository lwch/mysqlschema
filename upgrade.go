package mysqlschema

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"sort"
	"strconv"
	"strings"
)

func doUpgrade(db *sql.DB, name string, upgrade []string, exists *bool, current, class int) error {
	ver := func(dir string) int64 {
		dir = path.Base(dir)
		dir = strings.TrimPrefix(dir, "v")
		dir = strings.TrimSuffix(dir, ".sql")
		n, _ := strconv.ParseInt(dir, 10, 64)
		return n
	}
	sort.Slice(upgrade, func(i, j int) bool {
		return ver(upgrade[i]) < ver(upgrade[j])
	})
	run := func(dir string, current, upgrade int) (err error) {
		code, err := ioutil.ReadFile(dir)
		if err != nil {
			return err
		}
		if len(code) > 0 && !strings.Contains(string(code), name) {
			return fmt.Errorf("table name %s not found in %s", name, dir)
		}
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
				return
			}
			if !*exists {
				*exists = true
			}
		}()
		if len(code) > 0 {
			_, er = db.Exec(string(code))
			if er != nil {
				return er
			}
		}
		key := "schema_version"
		if class != 0 {
			key = "data_version"
		}
		if !*exists {
			_, er = db.Exec(fmt.Sprintf("INSERT INTO __tb_schema__(name, %s) VALUES(?, ?)", key),
				name, upgrade)
		} else {
			_, er = db.Exec(fmt.Sprintf("UPDATE __tb_schema__ SET %s=? WHERE name=?", key),
				upgrade, name)
		}
		if er != nil {
			return er
		}
		_, er = db.Exec(`INSERT INTO __tb_schema_logs__(name, class, current_version,
			upgrade_version, code) VALUES(?, ?, ?, ?, ?)`, name, class, current, upgrade, string(code))
		return er
	}
	key := "schema"
	if class != 0 {
		key = "data"
	}
	for _, file := range upgrade {
		nextVersion := ver(file)
		log.Printf("  * %s: upgrade [%s] from %d to %d...", key, name, current, nextVersion)
		err := run(file, current, int(nextVersion))
		if err != nil {
			return err
		}
		current = int(nextVersion)
	}
	return nil
}
