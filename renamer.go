/*
* renamer.go
* Rename downloaded torrents.
 */

package main

import (
    "fmt"
    "os"
    "time"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

func Renamer() {
    for {
        dbpath := Cachedir + "/" + Database
        db, err := sql.Open("sqlite3", dbpath)
        if err != nil {
            fmt.Println(err)
        }

        rows, err := db.Query(`SELECT * FROM ptp WHERE status="downloading"`)
        if err != nil {
            fmt.Println(err)
        }

        done := []int{}

        for rows.Next() {
            var id int
            var torrentname string
            var filename string
            var newfilename string
            var filelength int64
            var status string
            rows.Scan(&id, &torrentname, &filename, &newfilename, &filelength, &status)
            f, err := os.Open(Torrentdownload + "/" + filename)
            if err != nil {
                fmt.Println(err)
            }
            fi, err := f.Stat()
            if err != nil {
                fmt.Println(err)
            }
            currentfilesize := fi.Size()
            f.Close()
            if filelength == currentfilesize {
                fmt.Println("Copying completed files to final directory...")
                copyFileContents(Torrentdownload + "/" + filename, Torrentdst + "/" + newfilename)
                done = append(done, id)
            }
        }
        rows.Close()

        for _, id := range done {
            updateStatus(db, id)
        }

        db.Close()
        time.Sleep(60 * time.Second)
    }
}

func updateStatus(db *sql.DB, id int) {
    sql := fmt.Sprintf("UPDATE ptp SET status=\"complete\" WHERE id=%d", id)
    _, err := db.Exec(sql)
    if err != nil {
        fmt.Println(err)
    }
}