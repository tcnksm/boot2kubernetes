# boot2k8s

boot2k8s start single node [kubernetes](https://github.com/googlecloudplatform/kubernetes) cluster on local environment with _one command_ using docker. 

## Usage

```bash
$ boot2k8s up
```

## Install

To install, use `go get`:

```bash
$ go get -d github.com/tcnksm/boot2k8s
```

## Contribution

1. Fork ([https://github.com/tcnksm/boot2k8s/fork](https://github.com/tcnksm/boot2k8s/fork))
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
