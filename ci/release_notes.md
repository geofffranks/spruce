# Bug Fixes

- Specifying a static IP pool that ends before it starts no longer
  causes the `(( static_ips ... ))` operator to loop infiinitely
  trying to increment its way to the end of the range.
