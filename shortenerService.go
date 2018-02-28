package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"strconv"
	"io"
	"time"
	"net"
	"golang.org/x/net/netutil"
)

var shortener = &Shortener{}

func init() {
  	shortener.Init("./database.db")
}

func main() {

	defer shortener.Close()

	server := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1024,
	}

	http.HandleFunc("/set", func (w http.ResponseWriter, r *http.Request) {

		if r.Method == "OPTIONS" {
			header := w.Header()
			header.Set("Accept", "text/plain")
			header.Set("Allow", "OPTIONS, POST")
			header.Set("Cache-Control", "public, max-age=31536000")
			return
		}

		if r.Method != "POST" {
			header := w.Header()
			header.Set("Allow", "OPTIONS, POST")
			header.Set("Cache-Control", "public, max-age=31536000")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		if r.Header.Get("Content-Length") == "" {
			http.Error(w, "Content-Length wasn't specified", http.StatusBadRequest)
			return
		}

		contentLength, err := strconv.Atoi(r.Header.Get("Content-Length"))
		if  contentLength > 4096 {
			http.Error(w, "Max request content length is 4KB", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()
		buf := make([]byte, contentLength)
		_, err = io.ReadFull(r.Body, buf)

		if err == nil {

			longUrl := string(buf)
			_, err := url.Parse(longUrl)

			if err == nil {

				if strings.Index(longUrl, r.Host) != -1 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				shortUrl, err := shortener.Shorten(longUrl)
				if err == nil {
					header := w.Header()
					header.Set("Content-Type", "text/plain")
					w.Write([]byte(shortUrl.short))
					log.Printf("Saved url [%v]\n", shortUrl)
				} else {
					http.Error(w, fmt.Sprintf("Unable to get short url for [ %s ]", longUrl), http.StatusInternalServerError)
				}

			} else {
				http.Error(w, fmt.Sprintf("Given url wasn't properly formatted: %s", err.Error()), http.StatusInternalServerError)
			}

		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

	})

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {

		if r.Method == "OPTIONS" {
			header := w.Header()
			header.Set("Accept", "text/plain")
			header.Set("Allow", "OPTIONS, GET")
			header.Set("Cache-Control", "public, max-age=31536000")
			return
		}

		if r.Method != "GET" {
			header := w.Header()
			header.Set("Allow", "OPTIONS, GET")
			header.Set("Cache-Control", "public, max-age=31536000")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		url, err := shortener.Lookup(r.URL.Path[1:])
		if err == nil {
			log.Printf("Result was [%s]\n", url)
			http.Redirect(w, r, url.long, http.StatusMovedPermanently)
		} else {
			log.Printf("Result was [%s]\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}

	})

	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer listener.Close()

	listener = netutil.LimitListener(listener, 100)
	log.Fatal(server.Serve(listener))

}
