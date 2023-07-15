build:
	go build -o gio-cache ./main.go

override-static-files:
	cp ./static/* ./bin/
