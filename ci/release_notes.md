# New Features

- `(( load "FILENAME" ))` has been added. The `load` operator will load the contents of a YAML file,
  into the datastructure where it was invoked. See the [docs|https://github.com/geofffranks/spruce/blob/master/doc/operators.md#-load-]
- `(( vault ... ))` now supports the KV v2 API if you set `VAULT_VERSION=2` in your environment. By default, `spruce`
  will continue to use the v1 API (`VAULT_VERSION=1`)

# Acknowledgements

Many thanks to @HeavyWombat for the `load` operator!
Great work from @yurinnick to add the Vault KV v2 API support!
