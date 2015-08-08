# To use libcompose Godepped packages, docker/autogen/dockerversion
GOPATH_ := $(GOPATH)/src/github.com/docker/libcompose/Godeps/_workspace:$(GOPATH)

default: test

deps:
	go get -v golang.org/x/tools/cmd/vet	
	go get -v github.com/golang/lint/golint
	env GOPATH=$(GOPATH_) go get -d -v ./...

build: deps	
	env GOPATH=$(GOPATH_) go build -o bin/boot2k8s

install: deps
	env GOPATH=$(GOPATH_) go install

test: build
	go vet ./...
	env GOPATH=$(GOPATH_) go test -v ./...
