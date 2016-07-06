# Improvements

- Error messages from `(( param ... ))` operators are now
  de-duplicated from `(( grab ... ))` propagation, leading to more
  clear direction with heavily parameterized BOSH manifests.
  Fixes #129

- Added `(( insert ... ))` operator to insert array entries at
  a specific position in the target list. The operator behaves
  similar to `(( append ))` or `(( prepend ))`, but allows to
  specifiy either the desired insertion point by name
  (`(( insert after "nats" ))`) or by index (`(( insert after 4 ))`).

- The `(( append ))`,  `(( prepend ))`, and `(( insert ... ))`
  operators are now allowed multiple times in a list. This provides the
  flexibility to append, prepend, or insert elements using only one list:
  ```yml
  list:
  - (( prepend ))
  - The new first entry
  - (( append ))
  - The new last entry
  ``

- Updated godep to use golang vendoring structure, in preparation for
  removal of `Godeps/_workspace` in golang 1.8.
