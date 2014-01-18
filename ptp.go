/*
* ptp.go
* Login and fetch data from PTP API.
*/

package main

import (
    "strings"
    "errors"
    "net/url"
    "net/http"
    "net/http/cookiejar"
    "code.google.com/p/go.net/publicsuffix"
    "io/ioutil"
    "encoding/json"
)

const (
    ptp_endpoint = "https://tls.passthepopcorn.me"
)

type PTP struct {
    username, password, passkey, authkey string
    cookiejar http.CookieJar
}

type loginResult struct {
    Result string
}

func NewPTP(username, password, passkey string) PTP {
    ptp := PTP { 
        username: username,
        password: password,
        passkey: passkey,
    }
    return ptp
}

func (p *PTP) CheckLogin() (bool, error) {
    ptp_url, err := url.Parse(ptp_endpoint)
    if err != nil {
        return false, err
    }
    if p.cookiejar == nil || len(p.cookiejar.Cookies(ptp_url)) < 3 {
        return false, nil
    }
    return true, nil
}

func (p *PTP) Login() error {
    options := cookiejar.Options {
        PublicSuffixList: publicsuffix.List,
    }
    var err error
    p.cookiejar, err = cookiejar.New(&options)
    if err != nil {
        return err
    }

    client := &http.Client{ Jar: p.cookiejar }
    postData := url.Values { "username": {p.username}, 
    "password": {p.password}, "passkey": {p.passkey}, "keeplogged": {"1"} }
    resp, err := client.PostForm(ptp_endpoint + "/ajax.php?action=login", 
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

func (p *PTP) Get(imdbID string) ([]byte, error) {
    client := &http.Client { Jar: p.cookiejar }
    queryValues := url.Values { "imdb": {imdbID}, "json": {"1"} }
    req, err := http.NewRequest("GET", ptp_endpoint + "/torrents.php?" + 
        queryValues.Encode(), nil)
    //req, err := http.NewRequest("GET", "http://paste.ee/r/xIeue", nil)
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
        contents = []byte("{ \"Result\": \"Not found\" }")
    }

    return contents, nil
}