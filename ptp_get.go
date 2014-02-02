/*
* ptp_get.go
* Download torrent from PTP and add torrent info to struct.
 */

package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/huangbong/gotrntmetainfoparser"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type PTPGet struct {
	Cookiejar  http.CookieJar
	id         string
	authkey    string
	passkey    string
	title      string
	year       string
	source     string
	resolution string
	codec      string
	container  string
}

func NewPTPGet(cookiejar http.CookieJar, id, authkey, passkey string,
	title, year, source, resolution, codec, container string) PTPGet {
	ptp_get := PTPGet{
		Cookiejar:  cookiejar,
		id:         id,
		authkey:    authkey,
		passkey:    passkey,
		title:      title,
		year:       year,
		source:     source,
		resolution: resolution,
		codec:      codec,
		container:  container,
	}
	return ptp_get
}

func (p *PTPGet) Download() ([]byte, error) {
	client := &http.Client{Jar: p.Cookiejar}
	queryValues := url.Values{"action": {"download"}, "id": {p.id},
		"authkey": {p.authkey}, "passkey": {p.passkey}}
	req, err := http.NewRequest("GET", ptp_endpoint_tls+"/torrents.php?"+
		queryValues.Encode(), nil)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	out, err := ioutil.TempFile(Cachedir+"/ptp/", "ptp_")
	if err != nil {
		return nil, err
	}
	defer out.Close()

	if _, err := out.Write(contents); err != nil {
		return nil, err
	}

	if err := p.Parse(out.Name()); err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf("{\"Result\":\"%s\"}", "Ok")), nil
}

func (p *PTPGet) Parse(torrentname string) error {
	fmt.Println("Parsing... " + torrentname)

	var metaInfo gotrntmetainfoparser.MetaInfo
	if result := metaInfo.ReadTorrentMetaInfoFile(torrentname); result == false {
		return errors.New("Could not parse torrent file.")
	}
	metaInfo.DumpTorrentMetaInfo()

	var filename string
	var filelength int64

	switch len(metaInfo.Info.Files) {
	case 0:
		filename = metaInfo.Info.Name
		filelength = metaInfo.Info.Length
	case 1:
		filename = metaInfo.Info.Files[0].Path[0]
		filelength = metaInfo.Info.Files[0].Length
	default:
		var largest int
		largest = 0
		for i := 0; i < len(metaInfo.Info.Files); i++ {
			if metaInfo.Info.Files[i].Length > metaInfo.Info.Files[largest].Length {
				largest = i
			}
		}
		filename = metaInfo.Info.Name + "/" + metaInfo.Info.Files[largest].Path[0]
		filelength = metaInfo.Info.Files[largest].Length
	}

	fileslice := []string{p.title, p.year, p.source, p.resolution, p.codec, p.container}

	newfilename := strings.ToLower(strings.Join(fileslice, "."))
	newfilename = strings.Replace(newfilename, " ", ".", -1)
	newfilename = strings.Replace(newfilename, "blu-ray", "bluray", -1)

	err := p.InsertTorrent(torrentname, filename, newfilename, filelength)
	if err != nil {
		return err
	}

	fmt.Println("Copying torrent file to watch directory...")
	err = p.CopyTorrentFile(torrentname)
	if err != nil {
		return err
	}

	return nil
}

func (p *PTPGet) InsertTorrent(torrentname, filename, newfilename string, filelength int64) error {
	dbpath := Cachedir + "/" + Database
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT into ptp(id, torrentname, filename, newfilename, filelength, status) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(nil, torrentname, filename, newfilename, filelength, "downloading")
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (p *PTPGet) CopyTorrentFile(torrentname string) error {
	torrentNameSplit := strings.Split(torrentname, "/")
	dst := torrentNameSplit[len(torrentNameSplit)-1] + ".torrent"
	dst = Torrentwatch + "/" + dst
	err := copyFileContents(torrentname, dst)
	if err != nil {
		return err
	}
	return nil
}

func copyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return nil
}
