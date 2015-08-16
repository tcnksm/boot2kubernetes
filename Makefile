# Embedding commit hash
COMMIT = $$(git describe --always)

# gox configuration
XC_OS := "darwin linux windows"
XC_ARCH := "386 amd64"

default: test

deps: clean
	go get -v golang.org/x/tools/cmd/vet	
	go get -v github.com/golang/lint/golint
	go get -v github.com/jteeuwen/go-bindata
	go get -v -d github.com/docker/libcompose
	go get -v -d github.com/docker/docker
	go get -v golang.org/x/net/html/atom
	go get -v golang.org/x/crypto/ssh
	mkdir $(GOPATH)/src/github.com/docker/docker/autogen
	ln -s $(GOPATH)/src/github.com/docker/libcompose/Godeps/_workspace/src/github.com/docker/docker/autogen/dockerversion $(GOPATH)/src/github.com/docker/docker/autogen/dockerversion
	go get -d -v ./... 

bindata: deps
	cd config && $(GOPATH)/bin/go-bindata -pkg="config" .	

build: bindata
	env GOPATH=$(GOPATH_) go build -o bin/boot2k8s

package: bindata
	@sh -c "'$(CURDIR)/scripts/package.sh'"

test: build
	go vet ./...
	go test -race
	env GOPATH=$(GOPATH_) go test -v ./...

dist-docker: bindata
	/usr/local/bin/docker run --rm -v $(GOPATH)/src/github.com/tcnksm/boot2kubernetes:/gopath/src/github.com/tcnksm/boot2kubernetes -w /gopath/src/github.com/tcnksm/boot2kubernetes tcnksm/gox:1.5rc make package

clean:
	rm -fr $(GOPATH)/src/github.com/docker/docker/autogen
