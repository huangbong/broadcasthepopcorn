/*
* broadcasthepopcorn.go
* PTP and BTN autodownloader and organizer.
*/

package main

import (
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
    cache_dir = "_cache"
    ptp_username = ""
    ptp_password = ""
    ptp_passkey = ""
)

type appError struct {
    Error error
    Message string
    Code int
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if e := fn(w, r); e != nil {
        log.Println(e.Error)
        http.Error(w, e.Message, e.Code)
    }
}

var cache Cache
var ptp PTP

func main() {
    // watch for SIGTERM
    go watch()

    // route dynamic URLs
    r := mux.NewRouter()
    r.Handle("/", appHandler(index_view))
    r.Handle("/movies", appHandler(movies_view))
    ptp = NewPTP(ptp_username, ptp_password, ptp_passkey)
    cache = NewImageCache(cache_dir)
    r.Handle("/ptp", appHandler(ptp_view))
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


func ptp_view(w http.ResponseWriter, r *http.Request) *appError {
    w.Header().Set("Content-Type", "application/json")
    imdbID, err := checkQuery("imdbID", r)
    if err != nil {
        return &appError{ err, "No URL argument passed.", 500}
    }
    if logged_in, _ := ptp.CheckLogin(); logged_in == false {
        if err := ptp.Login(); err != nil {
            return &appError{ err, "Could not login to PTP.", 500}
        }
    }
    json, err := ptp.Get(imdbID)
    if err != nil {
        return &appError{ err, "Could not retrieve movie information.", 500}
    }
    w.Write(json)
    return nil
}

func image_view(w http.ResponseWriter, r *http.Request) *appError {
    url, err := checkQuery("url", r)
    if err != nil {
        return &appError{ err, "No URL argument passed.", 500}
    }
    if i, err := cache.Get(url); err != nil {
        return &appError{ err, "Could not cache image.", 500 }
    } else {
        w.Write(i)
    }
    return nil
}

func viewTemplate(filename string, w http.ResponseWriter) *appError {
    t := template.New(filename)
    parse, err := t.ParseGlob("templates/*.html")
    if err != nil {
        return &appError{ err, "Template files not found.", 404 }
    }
    t = template.Must(parse, err)
    if err := t.Execute(w, nil); err != nil {
        return &appError{ err, "Could not load templates.", 500 }
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