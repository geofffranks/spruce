# New Features

- There is a new `(( ips ))` operator, provided by Soren Hansen!
  It calculates IPs using indices and offests from a given IP or
  CIDR block. Useful if you need to calculate an IP in a non-BOSH
  YAML file, or if the IPs fall outside of the range of BOSH networks
  defined in them anifest. See the [IPs operator docs][1] for more info.

# Improvements

- `spruce` is now built with golang 1.9

# Bug Fixes

- `spruce` will now honor the System's trusted root CAs on Darwin,
  when connecting to Vault

# Acknowledgements

Thanks @sorenh for your work on the `(( ips ))` operator, and the bugfixes you
provided for the 1.12.2 release!

[1]: https://github.com/geofffranks/spruce/blob/master/doc/operators.md#-ips-
