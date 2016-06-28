# Improvements

- Error messages from `(( param ... ))` oeprators are now
  de-duplicated from `(( grab ... ))` propagation, leading to more
  clear direction with heavily parameterized BOSH manifests.
  Fixes #129
