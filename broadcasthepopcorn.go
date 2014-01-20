/*
* broadcasthepopcorn.go
* PTP and BTN autodownloader and organizer.
*/

package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "log"
    "errors"
    "net/http"
    "html/template"
    "github.com/gorilla/mux"
)

const (
    cache_dir = ""
    ptp_username = ""
    ptp_password = ""
    ptp_passkey = ""
)

type appError struct {
    Error error
    Message string
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if e := fn(w, r); e != nil {
        log.Println(e.Error)
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(e.Message));
    }
}

var cache Cache
var ptp_search PTPSearch

func main() {
    // watch for SIGTERM
    go watch()

    // route dynamic URLs
    r := mux.NewRouter()
    r.Handle("/", appHandler(index_view))
    r.Handle("/movies", appHandler(movies_view))
    ptp_search = NewPTP(ptp_username, ptp_password, ptp_passkey)
    cache = NewImageCache(cache_dir)
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

func jsonResult(s string) (string) {
    json := fmt.Sprintf("{\"Result\":\"%s\"}", s)
    return json
}

func ptp_search_view(w http.ResponseWriter, r *http.Request) *appError {
    w.Header().Set("Content-Type", "application/json")
    imdbID, err := checkQuery("imdbID", r)
    if err != nil {
        return &appError{ err, jsonResult("No URL argument passed.") }
    }
    if logged_in, _ := ptp_search.CheckLogin(); logged_in == false {
        if err := ptp_search.Login(); err != nil {
            return &appError{ err, jsonResult("Could not login to PTP.") }
        }
    }
    json, err := ptp_search.Get(imdbID)
    if err != nil {
        return &appError{ err, jsonResult("Could not retrieve movie information.") }
    }
    w.Write(json)
    return nil
}

func image_view(w http.ResponseWriter, r *http.Request) *appError {
    url, err := checkQuery("url", r)
    if err != nil {
        return &appError{ err, jsonResult("No URL argument passed.") }
    }
    if i, err := cache.Get(url); err != nil {
        return &appError{ err, jsonResult("Could not cache image.") }
    } else {
        w.Write(i)
    }
    return nil
}

func viewTemplate(filename string, w http.ResponseWriter) *appError {
    t := template.New(filename)
    parse, err := t.ParseGlob("templates/*.html")
    if err != nil {
        return &appError{ err, jsonResult("Template files not found.") }
    }
    t = template.Must(parse, err)
    if err := t.Execute(w, nil); err != nil {
        return &appError{ err, jsonResult("Could not load templates.") }
    }
    return nil
}

func checkQuery(field string, r *http.Request) (string, error) {
    if err := r.ParseForm(); err != nil {
        return "", err
    }
    if len(r.Form[field]) == 0 {
        return "", errors.New("Error")
    }
    query := r.Form[field][0]
    return query, nil
}

func watch() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    signal.Notify(c, syscall.SIGTERM)
    <-c
    if err := os.RemoveAll(cache_dir); err != nil {
     log.Fatal(err)
    } else {
     log.Println("Successfully closed.")
     os.Exit(0)
    }
}