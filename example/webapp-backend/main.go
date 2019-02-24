//go:generate vault-cli -s -n react ../webapp-frontend/build ./resv2

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/pkg/browser"

	res "github.com/go-sharp/vault/example/webapp-backend/resv2"
)

func sayHelloHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("calling web api: %v...", r.URL.Path)
	name := r.FormValue("name")
	if name == "" {
		fmt.Fprintf(w, "Hello, anonymous user!")
		return
	}

	fmt.Fprintf(w, "Hello, %v!", name)
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("calling web api: %v...", r.URL.Path)
	fmt.Fprintf(w, "%v", time.Now().Format("Mon Jan 2 15:04:05"))
}

func main() {
	loader := res.NewReactLoader()

	http.HandleFunc("/api/sayhello", sayHelloHandler)
	http.HandleFunc("/api/time", timeHandler)

	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(loader)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fp := r.URL.Path
		if fp == "/" {
			fp = "/index.html"
		}
		log.Printf("requesting: %v...", r.URL.Path)

		f, err := loader.Open(fp)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		defer f.Close()

		fi, _ := f.Stat()
		w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(fi.Name())))
		w.Write(data)
	})

	log.Println("webapp started, listening on port :8080...")
	browser.OpenURL("http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
