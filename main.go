// gio-cache instructs the browser to cache your wasm app until you update the version or significant time has passed
// in addition it shows a simple static css spinner while the app is being loaded and initialized
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	origin                 = "*"
	methods                = "GET"
	cacheControl           = "no-cache"
	wasmBinarVersionEnvVar = "WASM_BINARY_VERSION"
	port                   = "5173"
	staticDir              = "bin"

	commonWASMPath = "/main.wasm"
	gzipWASMPath   = "/main.wasm.gz"
	brotliWASMPath = "/main.wasm.br"
	zstdWASMPath   = "/main.wasm.zst"
)

type Binary struct {
	Path, Directive string
	Bytes           []byte
	Valid           bool
}

var (
	defaultBin = &Binary{Path: commonWASMPath, Directive: "identity"}
	binaries   = []*Binary{
		{Path: brotliWASMPath, Directive: "br"},
		{Path: zstdWASMPath, Directive: "zstd"},
		{Path: gzipWASMPath, Directive: "gzip"},
		defaultBin,
	}
)

func getMatchingBinary(acceptEncodingHeader string) *Binary {
	for _, bin := range binaries {
		for _, v := range strings.Split(acceptEncodingHeader, ",") {
			v = strings.TrimSpace(v)
			if len(v) <= 0 {
				continue
			}
			ve := strings.Split(v, ";")
			if len(ve) <= 0 {
				continue
			}

			if bin.Directive == ve[0] && bin.Valid {
				return bin
			}
		}
	}
	return defaultBin
}

func gioCache() error {
	var (
		err error
	)
	for _, bin := range binaries {
		var fl *os.File
		fl, err = os.Open(filepath.Join(staticDir, bin.Path))
		if err == nil {
			bin.Bytes, err = io.ReadAll(fl)
			if err == nil {
				bin.Valid = true
			}
		}
	}
	var (
		currentVersion = os.Getenv(wasmBinarVersionEnvVar)
		fileServer     = http.FileServer(http.Dir(staticDir))
	)
	if currentVersion == "" {
		return fmt.Errorf("You must set the env var: %s", wasmBinarVersionEnvVar)
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

		if r.URL.Path == commonWASMPath {
			var matchedBin = getMatchingBinary(r.Header.Get("Accept-Encoding"))

			w.Header().Set("Content-Type", "application/wasm")
			w.Header().Add("Content-Encoding", matchedBin.Directive)
			http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(matchedBin.Bytes))
			yellow("SERVED", fmt.Sprintf("%s %s %s", r.URL.Path, cacheDiff, matchedBin.Directive))
			return
		}
		yellow("SERVED", fmt.Sprintf("%s %s", r.URL.Path, cacheDiff))

		fileServer.ServeHTTP(w, r)
	})

	log.Println("Serving WASM app on port ", port)
	http.ListenAndServe(":"+port, nil)
	return nil
}

func main() {
	if err := gioCache(); err != nil {
		panic(err)
	}
}
