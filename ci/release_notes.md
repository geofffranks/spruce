# New Features

- `spruce` is now split into a main spruce package, a logging package, and a cmd package.
  This allows people to easily embed spruce's merging logic inside other Go applications.
  Unfortunately, to retrieve the binary via `go get` now, you will need to `go get github.com/geofffranks/spruce/...`.
  However, homebrew, and downloading binaries directly from https://github.com/geofffranks/spruce/releases still
  work as they used to.

# Bug Fixes

- `spruce` now supports BOSH 2.0 manifests with regards to the `static_ips` operator
