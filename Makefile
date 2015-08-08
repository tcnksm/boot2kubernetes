default: test

deps:
	go get -v golang.org/x/tools/cmd/vet	
	go get -v github.com/golang/lint/golint
	go get -v -d -t ./...

build: deps
	go build -o bin/boot2k8s

install: deps
	go install

test: build
	go test -v ./...
