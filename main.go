// gio-cache instructs the browser to cache your wasm app until you update the version or significant time has passed
// in addition it shows a simple static css spinner while the app is being loaded and initialized
package main

import (
	"log"
	"net/http"
	"os"
	"sync"
)

const (
	origin                 = "*"
	methods                = "GET"
	cacheControl           = "public, max-age=31536000"
	wasmBinarVersionEnvVar = "WASM_BINARY_VERSION"
	port                   = "5173"
	staticDir              = "bin"
)

var (
	timesServed uint64
	// currentVersion = "v0.0.14"
	mu sync.Mutex
)

func incTimesServed() {
	mu.Lock()
	defer mu.Unlock()
	timesServed++
}

func getTimesServed() uint64 {
	mu.Lock()
	defer mu.Unlock()
	return timesServed
}

func main() {
	wasmBinaryServer := http.FileServer(http.Dir(staticDir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var currentVersion = os.Getenv(wasmBinarVersionEnvVar)
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", methods)

		w.Header().Set("Cache-Control", cacheControl)
		w.Header().Set("ETag", currentVersion)
		if match := r.Header.Get("If-None-Match"); match != "" {
			if match == currentVersion {
				log.Println("[CACHED]")
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
		incTimesServed()
		log.Println("[UPDATING] ", getTimesServed())

		wasmBinaryServer.ServeHTTP(w, r)
	})

	log.Println("Serving WASM app on port ", port)
	http.ListenAndServe(":"+port, nil)
}
