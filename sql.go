/*
* sql.go
* Generate database at launch.
 */

package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func NewDatabase(filename string) error {
	if result, _ := exists(Cachedir); result == false {
		os.MkdirAll(Cachedir+"/imdb/", 0777)
		os.MkdirAll(Cachedir+"/ptp/", 0777)
	}

	dbpath := Cachedir + "/" + filename
	if result, _ := exists(dbpath); result == true {
		return nil
	}

	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return err
	}
	defer db.Close()

	var sql string

	sql = `
	CREATE TABLE "imdbcache" (
		"id" INTEGER PRIMARY KEY,
		"url" INTEGER NOT NULL,
		"filename" TEXT NOT NULL
	);
	`

	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	sql = `
	CREATE TABLE "ptp" (
		"id" INTEGER PRIMARY KEY,
		"torrentname" INTEGER NOT NULL,
		"filename" TEXT NOT NULL,
		"newfilename" TEXT NOT NULL,
		"filelength" INTEGER NOT NULL,
		"status" TEXT NOT NULL
	);
	`

	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	return nil

}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
