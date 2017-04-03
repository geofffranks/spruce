# Improvements
- The `(( vault ))` operator now caches data it has
  looked up, to improve performance on large-scale credential
  retrieval

# Bug Fixes

- Resolved issue #205 causing spruce to panic when merging
  arrays via a common key, and that key was an array/map

# Acknowledgements

Many thanks @thomasmmitchell for all the code in this release!
