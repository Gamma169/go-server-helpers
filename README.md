# go-server-helpers


## Package Changes

**NOTE:**  This package is a work in progress and subject to change.

Before v1.0.0 will use minor version tags to mark if the package exposed functions have changed.

**Minor Version Update:** If library goes from `v0.1.x` => `v0.2.x`, it means some function signatures have changed and **MAY NOT BE COMPATIBLE WITH PREVIOUS MINOR VERSION**

**Patch Version Update:** If library goes from `v0.2.3` => `v0.2.4`, it means new functions were added or bugs were fixed.  Library **WILL** be backwards-compatible with same minor version.


#### One-liner that does everything

```
gofmt -w . && go get ./... && go build ./... && go test -v ./tests
```
