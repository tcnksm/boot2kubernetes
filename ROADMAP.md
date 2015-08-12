# ROADMAP

This document shows which we will do for next release (TODO) or future plan.

## 0.1.0

- (**DONE**) `destroy` command to stop all related container
- [mitchellh/panicwrap](https://github.com/mitchellh/panicwrap) for crash reporting
- Fix panic when started from image pulling
- Fix bug sometimes it can not start port forwarding, because contaienr is not ready
- Binary release (and homebrew formula)

## Future

- `upgrade` to replace command to new one (Check how boot2docker does it)
- Integrate docker-machine to setup docker environment not only local but some cloud provider and start k8s there
- Enable to change kubernetes version (multiple `k8s.yml` ?)