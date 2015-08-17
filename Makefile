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

# Delete directories for using build
clean: 
	rm -fr $(TMP_GOPATH)
	rm -fr bin

test-deps:
	go get -v golang.org/x/tools/cmd/vet
	go get -v github.com/golang/lint/golint

deps: clean
	go get -v github.com/jteeuwen/go-bindata/...
	go get -v -d github.com/docker/libcompose
	@go get golang.org/x/net/html golang.org/x/oauth2 # FIXME
	@mkdir $(TMP_GOPATH)
	@GOPATH=$(TMP_GOPATH) go get golang.org/x/crypto/ssh 
	GOPATH=$(GOPATH_) go get -d -v ./... 

bindata: deps
	cd config && $(GOPATH)/bin/go-bindata -pkg="config" .	

build: bindata
	GOPATH=$(GOPATH_) go build -o bin/boot2k8s -ldflags "-X main.GitCommit \"$(COMMIT)\""

release: bindata
	go get -v github.com/tcnksm/ghr # For parallel uploading
	sh -c "'$(CURDIR)/scripts/release.sh' $(GITHUB_TOKEN)"

test: test-deps build
	go vet ./...
	go test -race
	go test -v ./...

release-docker:
	/usr/local/bin/docker run --rm -v $(REPO_PATH):/gopath/src/$(REPO) -w /gopath/src/$(REPO) -e GITHUB_TOKEN=$(GITHUB_TOKEN) tcnksm/gox:1.4.2 sh -c "apt-get update -y && apt-get install --no-install-recommends zip && make release"
