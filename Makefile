ORG := github.com/tcnksm
REPO := $(ORG)/boot2kubernetes
REPO_PATH :=$(GOPATH)/src/$(REPO)

# WARN: libcompose includes partial/incomplete golang.org/x/crypto/ssh package as Godep dependency.
# So if boot2k8s includes ssh package, go build tries to read partial/incomplete ssh package and fails..
# libcompose vendors package which is not go-get-able, so it should be high priority GOPATH,
# To only use normal ssh package first, create TMP_GOPATH and set it as first GOPATH
TMP_GOPATH := $(REPO_PATH)/tmp_gopath
GOPATH_ := $(TMP_GOPATH):$(GOPATH)/src/github.com/docker/libcompose/Godeps/_workspace:$(GOPATH)

# Embedding commit hash
COMMIT = $$(git describe --always)

# gox configuration
XC_OS := "darwin linux windows"
XC_ARCH := "386 amd64"

default: test

# Delete tmp_gopath directory which create temporary gopath
clean: 
	rm -fr $(REPO_PATH)/tmp_gopath

deps: clean
	go get -v golang.org/x/tools/cmd/vet
	go get -v github.com/golang/lint/golint
	go get -v github.com/jteeuwen/go-bindata/...
	go get -v -d -u github.com/docker/libcompose
	go get -v golang.org/x/net/html
	go get -v golang.org/x/crypto/ssh
	go get -v golang.org/x/oauth2
	mkdir -p $(TMP_GOPATH)/src/golang.org/x/crypto
	ln -s $(GOPATH)/src/golang.org/x/crypto/ssh $(TMP_GOPATH)/src/golang.org/x/crypto/ssh
	GOPATH=$(GOPATH_) go get -d -v ./... 

bindata: deps
	cd config && $(GOPATH)/bin/go-bindata -pkg="config" .	

build: bindata
	GOPATH=$(GOPATH_) go build -o bin/boot2k8s -ldflags "-X main.GitCommit \"$(COMMIT)\""

package: bindata
	sh -c "'$(CURDIR)/scripts/package.sh'"

test: build
	go vet ./...
	go test -race
	go test -v ./...

build-docker:
	/usr/local/bin/docker run --rm -v $(REPO_PATH):/gopath/src/$(REPO) -w /gopath/src/$(REPO) tcnksm/gox:1.5rc make build

dist-docker:
	/usr/local/bin/docker run --rm -v $(REPO_PATH):/gopath/src/$(REPO) -w /gopath/src/$(REPO) tcnksm/gox:1.5rc make package
