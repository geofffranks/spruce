# Bug Fixes

- The `(( static_ips ))` call now properly evaluates when
  top level `networks`, `jobs`, or `instance_groups` keys are
  operators referencing other datastructures.

  Thanks @giner for reporting the issue!
