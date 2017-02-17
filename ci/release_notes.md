# New Features

- Experimental support for a `spruce diff` command that generates
  semantic differential reports highlighting the differences
  between two YAML files.  The differences are semantic, meaning
  that they do not care about whitespace, flow-vs-block, or map
  key ordering.  The produced diffs are stable, so that if you run
  `spruce diff` twice, without changing input files, you get the
  same output -- this is surprisingly helpful when trying to
  reconcile to large YAML documents.

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
