/*
* ptp.go
* Login and fetch data from PTP API.
 */

package main

import (
	"code.google.com/p/go.net/publicsuffix"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

const (
	ptp_endpoint = "https://tls.passthepopcorn.me"
)

type PTPSearch struct {
	username, password, passkey, authkey string
	movie_source, movie_resolution       string
	cookiejar                            http.CookieJar
}

type loginResult struct {
	Result string
}

type PTPJson struct {
	Page     string `json:"Page"`
	Result   string `json:"Result"`
	Groupid  string `json:"GroupId"`
	Authkey  string `json:"AuthKey"`
	Passkey  string `json:"PassKey"`
	Imdbid   string `json:"ImdbID"`
	Torrents []struct {
		Id            string `json:"Id"`
		Quality       string `json:"Quality"`
		Source        string `json:"Source"`
		Container     string `json:"Container"`
		Codec         string `json:"Codec"`
		Resolution    string `json:"Resolution"`
		Size          string `json:"Size"`
		Scene         bool   `json:"Scene"`
		Uploadtime    string `json:"UploadTime"`
		Snatched      string `json:"Snatched"`
		Seeders       string `json:"Seeders"`
		Leechers      string `json:"Leechers"`
		Releasname    string `json:"ReleasName"`
		Checked       bool   `json:"Checked"`
		Goldenpopcorn bool   `json:"GoldenPopcorn"`
		Recommended   bool   `json:"Recommended"`
	} `json:"Torrents"`
}

func NewPTPSearch(username, password, passkey, movie_source, movie_resolution string) PTPSearch {
	ptp := PTPSearch{
		username:         username,
		password:         password,
		passkey:          passkey,
		movie_source:     movie_source,
		movie_resolution: movie_resolution,
	}
	return ptp
}

// Check PTPSearch.cookiejar to see if we have already logged in.

func (p *PTPSearch) CheckLogin() (bool, error) {
	ptp_url, err := url.Parse(ptp_endpoint)
	if err != nil {
		return false, err
	}
	if p.cookiejar == nil || len(p.cookiejar.Cookies(ptp_url)) < 3 {
		return false, nil
	}
	return true, nil
}

// Login to PTP and save cookie in cookiejar.

func (p *PTPSearch) Login() error {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	var err error
	p.cookiejar, err = cookiejar.New(&options)
	if err != nil {
		return err
	}

	client := &http.Client{Jar: p.cookiejar}
	postData := url.Values{"username": {p.username},
		"password": {p.password}, "passkey": {p.passkey}, "keeplogged": {"1"}}
	resp, err := client.PostForm(ptp_endpoint+"/ajax.php?action=login",
		postData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result loginResult
	if err := json.Unmarshal(contents, &result); err != nil {
		return err
	}
	if result.Result != "Ok" {
		return errors.New("Could not login to PTP.")
	}
	return nil
}

// Get PTP JSON search result from imdbID.

func (p *PTPSearch) Get(imdbID string) ([]byte, error) {
	client := &http.Client{Jar: p.cookiejar}
	queryValues := url.Values{"imdb": {imdbID}, "json": {"1"}}
	req, err := http.NewRequest("GET", ptp_endpoint+"/torrents.php?"+
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

	if strings.Contains(string(contents), "html") {
		contents = []byte("{\"Result\":\"Movie not found on PTP.\"}")
		return contents, nil
	}

	var default_response PTPJson
	if err := json.Unmarshal(contents, &default_response); err != nil {
		return nil, err
	}

	var rank PTPJson
	if err := json.Unmarshal(contents, &rank); err != nil {
		return nil, err
	}

	p.Recommend(&default_response, &rank)

	response_byte, _ := json.Marshal(default_response)

	return response_byte, nil
}

// Determine recommended torrent to download.
// Very simple/broken algorithm at this point.

func (p *PTPSearch) Recommend(default_response, rank *PTPJson) {
	// re-order rank (type PTPJson) by most snatched

	for i := 0; i < len(rank.Torrents); i++ {
		max, _ := strconv.Atoi(rank.Torrents[i].Snatched)
		max_id := i
		for j := i + 1; j < len(rank.Torrents); j++ {
			if val, _ := strconv.Atoi(rank.Torrents[j].Snatched); val > max {
				max, _ = strconv.Atoi(rank.Torrents[j].Snatched)
				max_id = j
			}
		}
		rank.Torrents[i], rank.Torrents[max_id] =
			rank.Torrents[max_id], rank.Torrents[i]
	}

	var recommendedID string

	// perfect match

	if recommendedID == "" {
		for i := 0; i < len(rank.Torrents); i++ {
			t := rank.Torrents[i]
			if t.Source == p.movie_source &&
				t.Resolution == p.movie_resolution {
				recommendedID = t.Id
				break
			}
		}
	}

	// ignore source

	if recommendedID == "" {
		for i := 0; i < len(rank.Torrents); i++ {
			t := rank.Torrents[i]
			if t.Resolution == p.movie_resolution {
				recommendedID = t.Id
				break
			}
		}
	}

	// ignore everything, recommend most snatched

	if recommendedID == "" {
		for i := 0; i < len(rank.Torrents); i++ {
			t := rank.Torrents[i]
			recommendedID = t.Id
			break
		}
	}

	for i := 0; i < len(default_response.Torrents); i++ {
		if default_response.Torrents[i].Id == recommendedID {
			default_response.Torrents[i].Recommended = true
		}
	}
}
