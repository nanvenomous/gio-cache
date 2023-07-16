build:
	go build -o gio-cache ./...

setup:
	cp ./static/* ./bin/
	brotli -k -Z -f bin/main.wasm
	rm bin/main.wasm
	sed -i 's/main\.wasm/main\.wasm\.br/g' bin/wasm.js 
