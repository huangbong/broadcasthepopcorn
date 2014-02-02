/*
* cache.go
* Cache and store images from IMDb.
 */

package main

import (
	"github.com/nfnt/resize"
	"image/jpeg"
	"io/ioutil"
	"net/http"
)

type Cache interface {
	Get(url string) ([]byte, error)
}

type ImageCache struct {
	urls      map[string]string
}

func NewImageCache() ImageCache {
	i := ImageCache{
		urls:      make(map[string]string),
	}
	return i
}

func (i ImageCache) Get(url string) ([]byte, error) {
	if tmp_name, ok := i.urls[url]; ok {
		return i.getCachedImage(tmp_name)
	} else {
		return i.cacheImage(url)
	}
}

func (i ImageCache) getCachedImage(tmp_name string) ([]byte, error) {
	file, err := ioutil.ReadFile(tmp_name)
	if err != nil {
		return nil, err
	}
	return file, err
}

func (i ImageCache) cacheImage(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	Image, err := jpeg.Decode(res.Body)
	if err != nil {
		return nil, err
	}
	m := resize.Resize(160, 238, Image, resize.Lanczos3)

	out, err := ioutil.TempFile(Cachedir + "/imdb/", "imdb_")
	if err != nil {
		return nil, err
	}
	defer out.Close()

	jpeg.Encode(out, m, nil)

	i.urls[url] = out.Name()

	return i.getCachedImage(i.urls[url])
}