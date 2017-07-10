# Bug Fixes

- Fixes #201 by supporting `azs` key in subnets

  Previously only `az` was supported. Now when specifying
  multiple AZs in a subnet, all IPs from that subnet can be
  used in any instance-group/job that is in a zone that the
  subnet mentioned.

  This can lead to interesting scenarios when using mixes of
  multi-az subnets and single-az subnets, where different offsets
  can mean the same IP in a different zone, or the same index could
  mean different IPs in different zones. Try not to do this, as it will
  likely lead to confusion down the road. However, care is made to ensure
  that IPs are never re-used, regardless of what subnets/azs they were
  allowed to be used by.

  This should not affect any existing IP allocations, since previously the
  `azs` field wasn't looked at, and the old behaviors remain the same
  for `az` and no-azs.
