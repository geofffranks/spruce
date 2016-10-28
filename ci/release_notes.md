## Improvements

- Operator argument string literals can now contain `\n`, `\r` and
  `\t`, with the expected results.  Notably, this allows multiline
  string pasting, which was impossible in previous versions. (#175)

# New Features

- Added the `--cherry-pick <yaml.data.path>` flag, to allow easier pruning of
  everything except a specific path

  Thanks @HeavyWombat!

# Fixes

- The `spruce-darwin-amd64` binary attached to the release is now
  macOS Sierra compatible (it was built with golang 1.6.3)
