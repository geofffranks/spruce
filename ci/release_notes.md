# Improvements

- Error messages from missing `(( param ))` calls are now isolated,
  so that any errors that may be caused by the params not having been
  overridden is suppressed.
- Error messages are now colorized, if applicable, for easier readability
- Error messages resolve indexes out to names, if possible, for easier readability 
  (`jobs.0.networks.0.static_ips` becomes `jobs.myjob.networks.mynet.static_ips`)

# Bug Fixes
- Fixed a panic caused when the `(( inject ))` operator tried to do dependency
  checking on values that do not exist. This now results in an error message
  from the `(( inject ))` operator, instead of a panic.
