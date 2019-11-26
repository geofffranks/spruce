# Bug Fixes

- Resolved [a bug](https://github.com/geofffranks/spruce/issues/306) introduced in 1.24.0 where
  `spruce merge file1.yml -` failed to open `-` as STDIN. Thanks @MMedini for the report!
- `spruce merge` `spruce fan` and `spruce json` now have support for `-h/--help` flags. Thanks @MMedini
  for this report as well!
