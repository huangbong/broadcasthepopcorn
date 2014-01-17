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
    "net/http"
    "html/template"
    "github.com/gorilla/mux"
)

const (
    cacheDir = "cache"
)

var cache ImageCache

func main() {
    go Watch()
    r := mux.NewRouter()
    r.HandleFunc("/", Index)
    r.HandleFunc("/movies", Movies)
    cache = NewImageCacheStore(cacheDir)
    r.HandleFunc("/image", Image)
    http.Handle("/", r)
    http.Handle("/css/", http.FileServer(http.Dir("static")))
    http.Handle("/js/", http.FileServer(http.Dir("static")))
    http.Handle("/img/", http.FileServer(http.Dir("static")))
    http.Handle("/cache/", http.FileServer(http.Dir(".")))
    http.ListenAndServe(":8000", nil)
}

func Index(w http.ResponseWriter, r *http.Request) {
    t := template.New("index.html")
    t = template.Must(t.ParseGlob("templates/*.html"))
    t.Execute(w, nil)
}

func Movies(w http.ResponseWriter, r *http.Request) {
    t := template.New("movies.html")
    t = template.Must(t.ParseGlob("templates/*.html"))
    t.Execute(w, nil)
}

func Image(w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    url := r.Form["url"][0]
    if err != nil {
        log.Println(err)
    }
    i := cache.Get(&url)
    w.Write(i)
}

func Watch() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    signal.Notify(c, syscall.SIGTERM)
    <-c
    if err := os.RemoveAll(cacheDir); err != nil {
     log.Fatal(err)
    } else {
     log.Println("Successfully closed")
     os.Exit(0)
    }
}