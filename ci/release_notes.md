# Improvements

- When running `spruce json` while there was a key of integer 4, and a key of the
  string "4", with two different values, `spruce json` will only allow one of them
  to win, as JSON only permits strings as keys, and one will invevitably overwrite
  the other. This will now emit a warning by default. Thanks @sorenh for adding this!
- `spruce json` now supports an optional `--strict` flag to prevent non-string keys
  from being converted into json at all. Another thank you to @sorenh!

# Removals

- The long-deprecated `--concourse` flag has been removed from `spruce`. Please use
  either the `(( grab ))` operator, or Concourse's `((variable))` syntax for pulling
  in the credentials.
