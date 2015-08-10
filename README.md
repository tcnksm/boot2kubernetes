# boot2kubernetes

[![GitHub release](http://img.shields.io/github/release/tcnksm/boot2kubernetes.svg?style=flat-square)][release]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]

[release]: https://github.com/tcnksm/boot2kubernetes/releases
[license]: https://github.com/tcnksm/boot2kubernetes/blob/master/LICENSE
[godocs]: http://godoc.org/github.com/tcnksm/boot2kubernetes

`boot2k8s` start single node [kubernetes](https://github.com/googlecloudplatform/kubernetes) cluster with _**one command**_ using docker :whale:. The purpose of this projcet is building kubernetes environment in fast way for testing or experiment on your development enviromnet.

## Usage

To up cluster,

```bash
$ boot2k8s up
```

This command pulls required docker images and starts them.  After this, you can start to run `kubectl` (You need to install it). If you run docker on boot2docker-vm, it also starts port forwarding server for `kubectl`.

To down (stop) cluster,

```bash
$ boot2k8s down
```

## Install

To install, use `go get`:

```bash
$ go get -d github.com/tcnksm/boot2kubernetes
```

## Contribution

1. Fork ([https://github.com/tcnksm/boot2kubernetes/fork](https://github.com/tcnksm/boot2kubernetes/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## References

- [Running Kubernetes locally via Docker](https://github.com/GoogleCloudPlatform/kubernetes/blob/release-1.0/docs/getting-started-guides/docker.md)
- [1 command to Kubernetes with Docker compose](http://sebgoa.blogspot.jp/2015/04/1-command-to-kubernetes-with-docker.html)

## Author

[Taichi Nakashima](https://github.com/tcnksm)
