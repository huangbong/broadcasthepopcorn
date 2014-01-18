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
    r.Handle("/", appHandler(index))
    r.Handle("/movies", appHandler(movies))
    ptp = NewPTP(ptp_username, ptp_password, ptp_passkey)
    cache = NewImageCache(cache_dir)
    r.Handle("/ptp", appHandler(ptp_view))
    r.Handle("/image", appHandler(image))

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

func index(w http.ResponseWriter, r *http.Request) *appError {
    return viewTemplate("index.html", w)
}

func movies(w http.ResponseWriter, r *http.Request) *appError {
    return viewTemplate("movies.html", w)
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

func ptp_view(w http.ResponseWriter, r *http.Request) *appError {
    if err := r.ParseForm(); err != nil {
        return &appError{ err, "Could not parse form.", 500 }
    }
    if len(r.Form["imdbID"]) == 0 {
        return &appError{ nil, "No URL argument passed.", 500 }
    }
    imdbID := r.Form["imdbID"][0]
    if logged_in, _ := ptp.CheckLogin(); logged_in == false {
        if err := ptp.Login(); err != nil {
            return &appError{ err, "Could not login to PTP.", 500}
        }
    }
    results, err := ptp.Get(imdbID)
    if err != nil {
        return &appError{ err, "Could not retrieve movie information.", 500}
    }
    fmt.Fprintf(w, results)
    return nil
}

func image(w http.ResponseWriter, r *http.Request) *appError {
    if err := r.ParseForm(); err != nil {
        return &appError{ err, "Could not parse form.", 500 }
    }
    if len(r.Form["url"]) == 0 {
        return &appError{ nil, "No URL argument passed.", 500 }
    }
    url := r.Form["url"][0]
    if i, err := cache.Get(url); err != nil {
        return &appError{ err, "Could not cache image.", 500 }
    } else {
        w.Write(i)
    }
    return nil
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