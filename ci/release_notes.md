# New Features

- The `(( delete ))` operator can now delete strings form simple lists of
  strings. For example:

  ```
  --- # this list
  array:
  - first
  - second
  - third
  --- when merged with this
  array:
  - (( delete second ))
  --- # yields
  array:
  - first
  - third
  ```

  Many thanks to @qu1queee and @HeavyWombat for the new feature!

- The `(( vault ))` operator now honors the `HTTP_PROXY`, `HTTPS_PROXY`, and
  `NO_PROXY` environment variables for using a proxy to connect to the Vault server.

  Thank you @drnic for the update!
