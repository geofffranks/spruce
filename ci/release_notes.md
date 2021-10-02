# New Features

- Adds a new `(( negate <bool> ))` operator to flip a boolean, in order
  to help DRY up yaml configs needing boolean inversions.

- `(( stringify ))` now supports literals, to allow for things like `(( stringify this_optional_key || "default message" ))

- Bumps to golang 1.17.1, and updates dependencies

# Fixes

- Addresses an issue in the `(( static_ip ))` operator that could result in
  integer overload or wraparound if compiled on an architecture where golang's
  default size of an `int` was not `int64`.

# Acknowledgements

Thanks @oddbloke for the negate feature, and @isibeni for the stringify improvement!
