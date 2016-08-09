# New Features

- Added support for availability zone processing, in the
  `static_ips` operator. When using BOSH's AZ feature in
  your manifest, you can now do things like
  `static_ips: (( static_ips(0, "z2:2", "z1:3") ))` to indicate
  which IPs from which AZs you wish to use as static.
- You can delete elements of an array at merge time via
  `(( delete "key-name" ))` or `(( delete 1 ))`, to delete
  elements with the `name` value of `key-name`, or index of `1`.

# Fixes

- Added some missing examples when manipulating arrays.

# Notes

`spruce` is now updated for golang 1.6 and beyond. The vendored dependencies have been moved into `./vendor`,
and a Makefile was added to ease testing/vetting excluding vendored libraries where appropriate.

# Thank You!

Special thanks to @swisscom, @HeavyWombat, @JamesClonk, @qu1queee for all their hard work making
this release possible!
