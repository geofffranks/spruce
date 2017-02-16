# Improvements

- `spruce merge --cherry-pick` now no longer evaluates the entire
  tree of operators before returning a subset.  Instead, operators
  that are not involved in the final tree are ignored.

# Bug Fixes

- Resolved an issue with the `(( static_ips ))` operator not always resolving
  dependencies properly, in cases where there were networks without subnets,
  or networks without static IPs, in conjunction with networks that had them.

  e.g. You defined an internal network with static IPs, and a VIP network, with
  no subnets.
