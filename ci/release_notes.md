# Fixes

- Addresses an issue introduced in v1.8.4 where `VAULT_SKIP_VERIFY` was being
  overridden by the `~/.saferc` file. The new behavior is to skip verification
  if it was requested in either `VAULT_SKIP_VERIFY`, or the `~/.saferc` file requested
  it.
