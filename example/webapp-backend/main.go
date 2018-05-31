//go:generate vault-cli -s -n react ../webapp-frontend/build ./res

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-sharp/vault/example/webapp-backend/res"
)

func sayHelloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		fmt.Fprintf(w, "Hello, anonymous user!")
		return
	}

	fmt.Fprintf(w, "Hello, %v!", name)
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", time.Now().Format("Mon Jan 2 15:04:05"))
}

func main() {
	loader := res.NewReactLoader()

	http.HandleFunc("/api/sayhello", sayHelloHandler)
	http.HandleFunc("/api/time", timeHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fp := r.URL.Path
		if fp == "/" {
			fp = "/index.html"
		}
		log.Printf("requesting: %v...", r.URL.Path)

		f, err := loader.Load(fp)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		defer f.Close()

		w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(f.Name())))
		w.Write(data)
	})

	log.Println("webapp started, listening on port :8080...")
	http.ListenAndServe(":8080", nil)
}
