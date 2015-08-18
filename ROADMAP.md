# ROADMAP

This document shows which we will do for next release (TODO) or future plan.

## 0.1.0

- **DONE**: `destroy` command to stop all related container
- **DONE**: Fix bug sometimes it can not start port forwarding, because container is not ready
- **DONE**: Delete libcompose (logrus) output
- **DONE**: Binary release (and homebrew formula)
- **DONE**: `list` command to show all container which is started by boot2k8s

## Future

- [mitchellh/panicwrap](https://github.com/mitchellh/panicwrap) for crash reporting
- `upgrade` command to replace command to new one (Check how boot2docker does it)
- `clean` command to delete all related docker images
- Integrate docker-machine to setup docker environment not only local but some cloud provider and start k8s there
- Enable to change kubernetes version (multiple `k8s.yml` ?)
