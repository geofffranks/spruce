# New Features

- New `(( keys ... ))` operator for extracting the keys of a map
  into a list, elsewhere in your manifest.

# Bug Fixes

- Fixed Cursor.Glob aliasing bug.  This was a pretty severe bug
  affecting anyone using multiple static ranges in their network
  definition.
