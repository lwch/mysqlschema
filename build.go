package mysqlschema

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type table struct {
	name          string
	exists        bool
	schemaVersion int
	dataVersion   int
}

func Build(db *sql.DB, dir string, errExit bool) error {
	err := prepare(db)
	if err != nil {
		return err
	}

	info, err := loadInfo(db)
	if err != nil {
		return err
	}

	tables, err := filepath.Glob(path.Join(dir, "*"))
	if err != nil {
		return err
	}

	for _, table := range tables {
		fi, err := os.Stat(table)
		if err != nil {
			log.Printf("[WARN] stat for %s directory failed, err=%v", table, err)
			continue
		}
		if !fi.IsDir() {
			log.Printf("[WARN] directory %s is not directory, skiped", table)
			continue
		}
		err = buildDir(db, table, info[path.Base(table)])
		if err != nil {
			log.Printf("[ERROR] build schema for %s failed, err=%v", path.Base(table), err)
			if errExit {
				return err
			}
			continue
		}
	}
	return nil
}

func loadInfo(db *sql.DB) (map[string]table, error) {
	rows, err := db.Query("SELECT name, schema_version, data_version FROM __tb_schema__")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := make(map[string]table)
	for rows.Next() {
		var name string
		var schema, data int
		err = rows.Scan(&name, &schema, &data)
		if err != nil {
			return nil, err
		}
		ret[name] = table{
			name:          name,
			exists:        true,
			schemaVersion: schema,
			dataVersion:   data,
		}
	}
	return ret, nil
}

func buildDir(db *sql.DB, dir string, tb table) error {
	log.Printf("[INFO] upgrading table %s...", path.Base(dir))

	buildDir := path.Join(dir, "schema")
	fi, err := os.Stat(buildDir)
	if os.IsNotExist(err) || !fi.IsDir() {
		log.Println("[WARN] no schema found, skiped")
	} else {
		err = buildSchema(db, path.Base(dir), buildDir, &tb.exists, tb.schemaVersion)
		if err != nil {
			return err
		}
	}

	buildDir = path.Join(dir, "data")
	fi, err = os.Stat(buildDir)
	if os.IsNotExist(err) || !fi.IsDir() {
		log.Println("[WARN] no data found, skiped")
	} else {
		err = buildData(db, path.Base(dir), buildDir, &tb.exists, tb.dataVersion)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildSchema(db *sql.DB, name, dir string, exists *bool, current int) error {
	files, err := filepath.Glob(path.Join(dir, "v*.sql"))
	if err != nil {
		return err
	}
	var latest int
	var upgrade []string
	for _, file := range files {
		if path.Base(file) == "v0.sql" {
			continue
		}
		fileName := path.Base(file)
		fileName = strings.TrimPrefix(fileName, "v")
		fileName = strings.TrimSuffix(fileName, ".sql")
		ver, err := strconv.ParseInt(fileName, 10, 64)
		if err != nil {
			return fmt.Errorf("can not resolve version of file %s", file)
		}
		if ver > int64(latest) {
			latest = int(ver)
		}
		if ver > int64(current) {
			upgrade = append(upgrade, file)
		}
	}
	if !*exists {
		err = buildLatest(db, name, dir, latest)
		if err != nil {
			return err
		}
		*exists = true
		return nil
	}
	return doUpgrade(db, name, upgrade, exists, current, 0)
}

func buildData(db *sql.DB, name, dir string, exists *bool, current int) error {
	files, err := filepath.Glob(path.Join(dir, "v*.sql"))
	if err != nil {
		return err
	}
	var upgrade []string
	for _, file := range files {
		if path.Base(file) == "v0.sql" {
			continue
		}
		fileName := path.Base(file)
		fileName = strings.TrimPrefix(fileName, "v")
		fileName = strings.TrimSuffix(fileName, ".sql")
		ver, err := strconv.ParseInt(fileName, 10, 64)
		if err != nil {
			return fmt.Errorf("can not resolve version of file %s", file)
		}
		if ver > int64(current) {
			upgrade = append(upgrade, file)
		}
	}
	return doUpgrade(db, name, upgrade, exists, current, 1)
}
