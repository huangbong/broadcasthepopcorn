/*
* ptp_get.go
* Download torrent from PTP and add torrent info to struct.
 */

package main

import (
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"errors"
	"github.com/huangbong/gotrntmetainfoparser"
)

type PTPGet struct {
	cachedir string
	Cookiejar http.CookieJar
	id string
	authkey string
	passkey string
}

type PTPTorrent struct {
	name string
}

var torrents []PTPTorrent

func NewPTPGet(cookiejar http.CookieJar, cachedir, id, authkey, passkey string) PTPGet {
	ptp_get := PTPGet{
		Cookiejar: cookiejar,
		cachedir: cachedir,
		id: id,
		authkey: authkey,
		passkey: passkey,
	}
	return ptp_get
}

func (p *PTPGet) Download() error {
	client := &http.Client{Jar: p.Cookiejar}
	queryValues := url.Values{"action": {"download"}, "id": {p.id}, 
		"authkey": {p.authkey}, "passkey": {p.passkey}}
	req, err := http.NewRequest("GET", ptp_endpoint_tls +"/torrents.php?" +
		queryValues.Encode(), nil)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	out, err := ioutil.TempFile(p.cachedir, "ptp_")
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := out.Write(contents); err != nil {
		return err
	}

	if err := p.Parse(out.Name()); err != nil {
		return err
	}

	return nil
}

func (p *PTPGet) Parse(filename string) error {
	fmt.Println("parsing... " + filename)
	var metaInfo gotrntmetainfoparser.MetaInfo
	if result := metaInfo.ReadTorrentMetaInfoFile(filename); result == false {
		return errors.New("Could not parse torrent file.")
	}
	metaInfo.DumpTorrentMetaInfo()
	return nil
}