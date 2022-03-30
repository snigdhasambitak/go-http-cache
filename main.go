package main

import (
	"fmt"
	"github.com/snigdhasambitak/go-http-cache/cache"
	"github.com/snigdhasambitak/go-http-cache/cache/memory"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

var storage cache.Storage

func init() {
	storage = memory.NewStorage()
}

func main(){
	fmt.Println("Listening on 80...")
	http.HandleFunc("/withoutCache", serveStaticFileWithoutCache)
	http.Handle("/withCache", cached("10s", serveStaticFileWithCache))
	log.Fatal(http.ListenAndServe(":80", nil))
}

func serveStaticFileWithCache(w http.ResponseWriter, r *http.Request){
	time.Sleep(2 * time.Second)
	fmt.Println(r.URL.Path)
	p := "." + r.URL.Path
	if p == "./withCache" {
		p = "./static/withCache/index.html"
	}
	http.ServeFile(w,r,p)
}
func serveStaticFileWithoutCache(w http.ResponseWriter, r *http.Request){
	time.Sleep(2 * time.Second)
	fmt.Println(r.URL.Path)
	p := "." + r.URL.Path
	if p == "./withoutCache" {
		p = "./static/withoutCache/index.html"
	}
	http.ServeFile(w,r,p)
}
func cached(duration string, handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		content := storage.Get(r.RequestURI)
		if content != nil {
			fmt.Print("Cache Hit!\n")
			w.Write(content)
		} else {
			c := httptest.NewRecorder()
			handler(c, r)

			for k, v := range c.HeaderMap {
				w.Header()[k] = v
			}

			w.WriteHeader(c.Code)
			content := c.Body.Bytes()

			if d, err := time.ParseDuration(duration); err == nil {
				fmt.Printf("New page cached: %s for %s\n", r.RequestURI, duration)
				storage.Set(r.RequestURI, content, d)
			} else {
				fmt.Printf("Page not cached. err: %s\n", err)
			}

			w.Write(content)
		}

	})
}