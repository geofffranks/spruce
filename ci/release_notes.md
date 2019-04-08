# Bug Fixes

- Fixed `((calc))` to output integer values for its output, when
  the result is a large integer. Previously, it was treated as a float,
  and converted to scientific notation by the YAML marshaler. Fixes #286.

# Acknowledgements

Thanks @jhunt for finding and reporting the issue!
