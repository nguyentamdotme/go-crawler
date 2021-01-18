package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func DBConn() (db *sql.DB) {
    dbDriver := "mysql"
    dbUser := "root"
    dbPass := ""
    dbName := "go-crawler"
    db, err := sql.Open(dbDriver, dbUser + ":" + dbPass + "@tcp(localhost:3306)/" + dbName)
    if err != nil {
        panic(err.Error())
    }
    return db
}
