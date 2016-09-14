# Improvements

- Array operators are no longer required to be at zero-index
  of arrays they're defined in. Elements that come prior
  to an array merge operator are processed with the default merge
  behavior.
- Improved error messaging when elements are passed to
  the array delete operator.

# Fixes
- The `(( static_ips ))` operator has better dependency checking
  when elements inside the network definiton's static IP range is
  using a grab, or other operator.
- The `(( join ))` operator now has dependency checking, so it
  can resolve that its operands are resolved before joining them,
  when working with elements of lists.

#  Thanks

The entirety of this release has been brought to us by @thomasmmithcell
