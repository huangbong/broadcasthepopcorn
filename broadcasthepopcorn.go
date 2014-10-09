/*
* broadcasthepopcorn.go
* PTP and BTN autodownloader and organizer.
 */

package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

type appError struct {
	Error   error
	Message string
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil {
		log.Println(e.Error)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(e.Message))
	}
}

var s JSONSettings
var Cachedir string
var Database string
var Torrentwatch string
var Torrentdownload string
var Torrentdst string
var cache Cache
var ptp_search PTPSearch

func init() {
	s, err := NewJSONSettings()
	if err != nil {
		panic("Your settings.json file is not configured properly.")
	}
	Cachedir = s.Cachedir
	Database = s.Database
	Torrentwatch = s.Torrentwatch
	Torrentdownload = s.Torrentdownload
	Torrentdst = s.Torrentdst
	cache = NewImageCache()
	if err := NewDatabase(Database); err != nil {
		panic(err)
	}
	go Renamer()
	ptp_search = NewPTPSearch(s.PTP.Username, s.PTP.Password,
		s.PTP.Passkey, s.PTP.Settings.MovieSource,
		s.PTP.Settings.MovieResolution)
	if err := ptp_search.Login(); err != nil {
		panic(err)
	}
}

func main() {
	// route dynamic URLs
	r := mux.NewRouter()
	r.Handle("/", appHandler(index_view))
	r.Handle("/movies", appHandler(movies_view))
	r.Handle("/ptp_search", appHandler(ptp_search_view))
	r.Handle("/ptp_get", appHandler(ptp_get_view))
	r.Handle("/image", appHandler(image_view))

	// serve static files
	http.Handle("/css/", http.FileServer(http.Dir("static")))
	http.Handle("/js/", http.FileServer(http.Dir("static")))
	http.Handle("/img/", http.FileServer(http.Dir("static")))
	http.Handle("/cache/", http.FileServer(http.Dir(".")))

	// route gorilla/mux
	http.Handle("/", r)

	// run HTTP server
	http.ListenAndServe(":12345", nil)
}

func index_view(w http.ResponseWriter, r *http.Request) *appError {
	return viewTemplate("index.html", w, nil)
}

func movies_view(w http.ResponseWriter, r *http.Request) *appError {
	return viewTemplate("movies.html", w, nil)
}

func jsonResult(s string) string {
	json := fmt.Sprintf("{\"Result\":\"%s\"}", s)
	return json
}

func ptp_search_view(w http.ResponseWriter, r *http.Request) *appError {
	w.Header().Set("Content-Type", "application/json")
	query, err := checkQuery(r)
	imdbID := query["imdbID"][0]
	if err != nil {
		return &appError{err, jsonResult("No URL argument passed.")}
	}
	if logged_in, _ := ptp_search.CheckLogin(); logged_in == false {
		if err := ptp_search.Login(); err != nil {
			return &appError{err, jsonResult("Could not login to PTP.")}
		}
	}
	json, err := ptp_search.Get(imdbID)
	if err != nil {
		return &appError{err, jsonResult("Could not retrieve movie information.")}
	}
	w.Write(json)
	return nil
}

func ptp_get_view(w http.ResponseWriter, r *http.Request) *appError {
	w.Header().Set("Content-Type", "application/json")
	query, err := checkQuery(r)
	if err != nil {
		return &appError{err, jsonResult("No URL arguments passed.")}
	}
	var ptp_get PTPGet
	ptp_get = NewPTPGet(ptp_search.Cookiejar, query["id"][0], query["authkey"][0],
		query["passkey"][0], query["title"][0], query["year"][0],
		query["source"][0], query["resolution"][0], query["codec"][0],
		query["container"][0])
	result, err := ptp_get.Download()
	if err != nil {
		return &appError{err, jsonResult("Could not download torrent.")}
	}
	w.Write(result)
	return nil
}

func image_view(w http.ResponseWriter, r *http.Request) *appError {
	query, err := checkQuery(r)
	url := query["url"][0]
	if err != nil {
		return &appError{err, jsonResult("No URL arguments passed.")}
	}
	if i, err := cache.Get(url); err != nil {
		return &appError{err, jsonResult("Could not cache image.")}
	} else {
		w.Write(i)
	}
	return nil
}

func viewTemplate(filename string, w http.ResponseWriter, data interface{}) *appError {
	t := template.New(filename)
	parse, err := t.ParseGlob("templates/*.html")
	if err != nil {
		return &appError{err, jsonResult("Template files not found.")}
	}
	t = template.Must(parse, err)
	if err := t.Execute(w, data); err != nil {
		return &appError{err, jsonResult("Could not load templates.")}
	}
	return nil
}

func checkQuery(r *http.Request) (map[string][]string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	if len(r.Form) == 0 {
		return nil, errors.New("No URL arguments passed.")
	}
	query := r.Form
	return query, nil
}
