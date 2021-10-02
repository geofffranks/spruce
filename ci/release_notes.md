# New Features

- Adds a new `(( negate <bool> ))` operator to flip a boolean, in order
  to help DRY up yaml configs needing boolean inversions.

- `(( stringify ))` now supports literals, to allow for things like `(( stringify this_optional_key || "default message" ))

- Bumps to golang 1.17.1, and updates dependencies

# Acknowledgements

Thanks @oddbloke for the negate feature, and @isibeni for the stringify improvement!
