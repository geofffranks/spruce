# Bug Fixes

- Specifying a static IP pool that ends before it starts no longer
  causes the `(( static_ips ... ))` operator to loop infiinitely
  trying to increment its way to the end of the range.

- `spruce merge` now requires an explicit environment variable
   ($REDACT) to be set, or it will fail to render the final YAML
   if it contains any `(( vault ... ))` calls.  This should help
   cases where people accidentally forget to authenticate to the
   vault, and then deploy BOSH manifests that are REDACTED all
   over the place.
