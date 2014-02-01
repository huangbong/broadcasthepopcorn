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
	"os"
	"os/signal"
	"syscall"
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
var cache Cache
var ptp_search PTPSearch

func init() {
	s, err := NewJSONSettings()
	if err != nil {
		panic("Your settings.json file is not configured properly.")
	}
	ptp_search = NewPTPSearch(s.PTP.Username, s.PTP.Password,
		s.PTP.Passkey, s.PTP.Settings.MovieSource,
		s.PTP.Settings.MovieResolution)
	cache = NewImageCache(s.CacheDir)
}

func main() {
	// watch for SIGTERM
	go watch()

	// route dynamic URLs
	r := mux.NewRouter()
	r.Handle("/", appHandler(index_view))
	r.Handle("/movies", appHandler(movies_view))
	r.Handle("/ptp_search", appHandler(ptp_search_view))
	r.Handle("/image", appHandler(image_view))

	// serve static files
	http.Handle("/css/", http.FileServer(http.Dir("static")))
	http.Handle("/js/", http.FileServer(http.Dir("static")))
	http.Handle("/img/", http.FileServer(http.Dir("static")))
	http.Handle("/cache/", http.FileServer(http.Dir(".")))

	// route gorilla/mux
	http.Handle("/", r)

	// run HTTP server
	http.ListenAndServe(":8000", nil)
}

func index_view(w http.ResponseWriter, r *http.Request) *appError {
	return viewTemplate("index.html", w)
}

func movies_view(w http.ResponseWriter, r *http.Request) *appError {
	return viewTemplate("movies.html", w)
}

func jsonResult(s string) string {
	json := fmt.Sprintf("{\"Result\":\"%s\"}", s)
	return json
}

func ptp_search_view(w http.ResponseWriter, r *http.Request) *appError {
	w.Header().Set("Content-Type", "application/json")
	imdbID, err := checkQuery("imdbID", r)
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

func image_view(w http.ResponseWriter, r *http.Request) *appError {
	url, err := checkQuery("url", r)
	if err != nil {
		return &appError{err, jsonResult("No URL argument passed.")}
	}
	if i, err := cache.Get(url); err != nil {
		return &appError{err, jsonResult("Could not cache image.")}
	} else {
		w.Write(i)
	}
	return nil
}

func viewTemplate(filename string, w http.ResponseWriter) *appError {
	t := template.New(filename)
	parse, err := t.ParseGlob("templates/*.html")
	if err != nil {
		return &appError{err, jsonResult("Template files not found.")}
	}
	t = template.Must(parse, err)
	if err := t.Execute(w, nil); err != nil {
		return &appError{err, jsonResult("Could not load templates.")}
	}
	return nil
}

func checkQuery(field string, r *http.Request) (string, error) {
	if err := r.ParseForm(); err != nil {
		return "", err
	}
	if len(r.Form[field]) == 0 {
		return "", errors.New("No URL argument passed.")
	}
	query := r.Form[field][0]
	return query, nil
}

func watch() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	<-c
	if err := os.RemoveAll(s.CacheDir); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully closed.")
		os.Exit(0)
	}
}
