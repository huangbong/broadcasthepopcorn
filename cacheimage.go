/*
* cacheimage.go
* Cache and store images from IMDb.
*/

package main

import (
    "os"
    "log"
    "io/ioutil"
    "net/http"
    "image/jpeg"
    "github.com/nfnt/resize"
)

type ImageCache interface {
    Get(url *string) []byte
}

type ImageCacheStore struct {
    cacheDir string
    urls map[string]string
}

func NewImageCacheStore(cacheDir string) *ImageCacheStore {
    i := &ImageCacheStore{
        cacheDir: cacheDir,
        urls: make(map[string]string),
    }
    if result, _ := Exists(cacheDir); result == false {
        os.Mkdir(cacheDir, 0777)
    }
    return i
}

func (i *ImageCacheStore) Get(url *string) []byte {
    if tmpname, ok := i.urls[*url]; ok {
        return i.GetCachedImage(tmpname)
    } else {
        image := i.CacheImage(*url, i.cacheDir)
        return image
    }
}

func (i *ImageCacheStore) GetCachedImage(tmpname string) []byte {
    file, err := ioutil.ReadFile(tmpname)
    if err != nil {
        log.Fatal(err)
    }
    return file
}

func (i *ImageCacheStore) CacheImage(url, cacheDir string) []byte {
    res, err := http.Get(url)
    if err != nil {
        log.Fatal(err)
    }

    Image, err := jpeg.Decode(res.Body)
    if err != nil {
        log.Fatal(err)
    }

    m := resize.Resize(160, 238, Image, resize.Lanczos3)

    out, err := ioutil.TempFile(cacheDir, "imdb_")
    if err != nil {
        log.Fatal(err)
    }

    jpeg.Encode(out, m, nil)

    i.urls[url] = out.Name()

    out.Close()

    return i.GetCachedImage(i.urls[url])
}

func Exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}