# boot2kubernetes

[![GitHub release](http://img.shields.io/github/release/tcnksm/boot2kubernetes.svg?style=flat-square)][release]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]

[release]: https://github.com/tcnksm/boot2kubernetes/releases
[license]: https://github.com/tcnksm/boot2kubernetes/blob/master/LICENSE
[godocs]: http://godoc.org/github.com/tcnksm/boot2kubernetes

`boot2k8s` starts single node [kubernetes](https://github.com/googlecloudplatform/kubernetes) cluster in _**one command**_ using docker :whale:. The purpose of this project is building kubernetes in fast way for testing or experiment on your development environment. _Kubernetes version is 1.0.x_. 

## Usage

To up cluster,

```bash
$ boot2k8s up
```

This command pulls required docker images and starts them. You can check which docker image/option/command is used in [`k8s.yml`](/config/k8s.yml). After container is running, you can start to use `kubectl` (You need to install it by yourself). If you run docker on boot2docker-vm, it also starts port forwarding server to connect master APIs via local `kubectl`. 

To destroy cluster,

```bash
$ boot2k8s destroy
```

This command will destroy kubernetes containers started by `boot2k8s`. Not only that but also remove containers which are started by kubernetes (will ask confirmation). 

## Install

If you use OSX, you can use homebrew,

```bash
$ brew tap tcnksm/boot2k8s
$ brew install boot2k8s
```
If you are on other platform, download a binary from [release page](https://github.com/tcnksm/boot2kubernetes/releases) and place it on your `$PATH`.

## Contribution

1. Fork ([https://github.com/tcnksm/boot2kubernetes/fork](https://github.com/tcnksm/boot2kubernetes/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

To build `boot2k8s` from source, use `go get` and `make`, 

```bash
$ go get -d github.com/tcnksm/boot2kubernetes
$ cd $GOPATH/src/github.com/tcnksm/boot2kubernetes
$ make build
```

After this, binary is in `./bin` directory. 


## References

What `boot2k8s` does is same as official doc ["Running Kubernetes locally via Docker"](https://github.com/GoogleCloudPlatform/kubernetes/blob/release-1.0/docs/getting-started-guides/docker.md) describes. If you don't want to install additional fancy binary on your PC, should follow that article. I also inspired by an article ["1 command to Kubernetes with Docker compose"](http://sebgoa.blogspot.jp/2015/04/1-command-to-kubernetes-with-docker.html), thanks.

## Author

[Taichi Nakashima](https://github.com/tcnksm)
