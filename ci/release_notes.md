# New Features

- `(( prune ))` is a new operator that allows for specifying
  keys to be pruned inside your YAML files, similar to the `--prune`
  flag

# Fixes

- `spruce` is once again generated with CGO_ENABLED=0, to remove dynamically
  linked files, and increase portability. This regressed during a recent
  update to the CI pipeline.
