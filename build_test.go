package mysqlschema

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

const dbHost = "127.0.0.1"
const dbPort = 3306
const dbUser = "root"
const dbPass = "123456"
const dbName = "test"

func TestBuild(t *testing.T) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName))
	if err != nil {
		t.Fatalf("mysql connect failed: %v", err)
	}
	defer db.Close()
	db.Exec("DROP TABLE IF EXISTS __tb_schema__")
	db.Exec("DROP TABLE IF EXISTS __tb_schema_logs__")
	db.Exec("DROP TABLE IF EXISTS table1")
	err = Build(db, "testdata", true)
	if err != nil {
		t.Fatalf("build table schema failed: %v", err)
	}
}

func TestUpgrade(t *testing.T) {
	TestBuild(t)

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName))
	if err != nil {
		t.Fatalf("mysql connect failed: %v", err)
	}
	defer db.Close()
	err = Build(db, "testupgrade", true)
	if err != nil {
		t.Fatalf("build table schema failed: %v", err)
	}
}
