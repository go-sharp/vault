//go:generate vault-cli -s -n react ../webapp-frontend/build ./res

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pkg/browser"

	res "github.com/go-sharp/vault/example/webapp-backend/res"
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

	http.Handle("/", http.FileServer(loader))

	log.Println("webapp started, listening on port :8080...")
	browser.OpenURL("http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
