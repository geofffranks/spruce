# New Features

- Added a `(( empty hash/array/string ))` operator for emptying out values.
  Traditionally this could be done in the merge phase, by merging `null` or `~`
  on top of the value, and then merging an empty hash/array on top of that.
  Now you can do it in a single swipe. Vroom.

# Improvements

- New documentation for all the magical things you can now do with arrays!

# Bug Fixes

- Fixed a bug with the `(( prune ))` operator (#158)

# Thanks

Many thanks to @HeavyWombat and @thomasmmitchell for their work on this release!
