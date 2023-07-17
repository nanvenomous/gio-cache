build:
	go build -o gio-cache ./...

setup:
	cp ./static/* ./bin/
	brotli -k -Z -f bin/main.wasm
	zstd -z --ultra -22 bin/main.wasm
	gzip -k -9 bin/main.wasm
