// gio-cache instructs the browser to cache your wasm app until you update the version or significant time has passed
// in addition it shows a simple static css spinner while the app is being loaded and initialized
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"io"
	"strings"
)

const (
	origin                 = "*"
	methods                = "GET"
	cacheControl           = "no-cache"
	wasmBinarVersionEnvVar = "WASM_BINARY_VERSION"
	port                   = "5173"
	staticDir              = "bin"

	commonWASMPath   = "/main.wasm"
	gzipWASMPath   = "/main.wasm.gz"
	brotliWASMPath   = "/main.wasm.br"

	commonFile []byte
	gzipFile []byte
	brotliFile []byte
)

func main() {
	{
		f, err := os.Open(gzipWASMPath);
		if  err == nil {
			gzipFile, _ = io.ReadAll(f)
		}
	}

	{
		f, err := os.Open(brotliWASMPath);
		if  err == nil {
			brotliFile, _ = io.ReadAll(f)
		}
	}

	{
		f, err := os.Open(commonWASMPath);
		if  err == nil {
			commonFile, _ = io.ReadAll(f)
		}
	}
	
	var (
		currentVersion = os.Getenv(wasmBinarVersionEnvVar)
		fileServer     = http.FileServer(http.Dir(staticDir))
	)
	if currentVersion == "" {
		panic(fmt.Errorf("You must set the env var: %s", wasmBinarVersionEnvVar))
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", methods)
		w.Header().Set("ETag", currentVersion)
		w.Header().Set("Cache-Control", cacheControl)

		var match = r.Header.Get("If-None-Match")
		cacheDiff := fmt.Sprintf("%s -> %s", match, currentVersion)
		if match == currentVersion {
			green("CACHED", fmt.Sprintf("%s %s", r.URL.Path, cacheDiff))
			w.WriteHeader(http.StatusNotModified)
			return
		}
		yellow("SERVED", fmt.Sprintf("%s %s", r.URL.Path, cacheDiff))

		
		if r.URL.Path == commonWASMFile {

			content := commonFile
			compression := ""
			
			for _, v := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
				if len(v) <= 0 {
					continue
				}
				ve := strings.Split(v, ";")
				if len(ve) <= 0 {
					continue
				}
				switch ve[0] {
				case "br", " br":
					compression = "br"
					break
				case "gzip", " gzip":
					compression = "gzip"
				}
			}
		
			switch {
			case compression == "br" && brotliFile != nil:
				content = brotliWASMFile
			case compression == "gzip" && gzipFile != nil:
				content = gzipWASMFile
			}
		
			if compression != "" {
				h.Add("Content-Encoding", compression)
			}
		
			http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(content))
			return
		}

		fileServer.ServeHTTP(w, r)
	})

	log.Println("Serving WASM app on port ", port)
	http.ListenAndServe(":"+port, nil)
}
