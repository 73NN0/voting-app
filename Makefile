.PHONY: build test clean

build:
	go build -o bin/voting-app

test:
	go test -v ./...

clean:
	rm -rf bin/