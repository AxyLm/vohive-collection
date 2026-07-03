# VoHive Collection

This repository is a source snapshot collection for building and auditing
`github.com/iniwex5/vohive` with its related local dependencies.

## Contents

- `vohive`: main VoHive application.
- `quectel-qmi-go`: local copy of `github.com/iniwex5/quectel-qmi-go`.
- `netlink`: local copy of `github.com/iniwex5/netlink`.
- `qqbot`: local copy of `github.com/iniwex5/qqbot`.
- `euicc-go`: local copy of `github.com/damonto/euicc-go`.
- `uicc-go`: local copy of `github.com/damonto/uicc-go`.

`github.com/iniwex5/vowifi-go` is not included in this snapshot.

## Layout Notes

The directories are committed as ordinary source files in this collection
repository, not as Git submodules. The original nested Git metadata is kept only
in the local working copy and is not part of this repository.

`vohive/go.mod` still refers to upstream module paths. To build entirely from
these local source copies, add local `replace` directives or use a `go.work`
file.

## Runtime Data

Runtime caches such as `vohive/data/mcc-mnc-table.json` are not included here.
VoHive can download the MCC/MNC table at runtime when the cache is missing.
