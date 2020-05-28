# Fixes

- `spruce` no longer errors out when it is passed an empty (null) document to merge. It
  instead treats it as an empty map.

- `spruce diff` would panic under certain circumstnaces outlined in [issue #318](https://github.com/geofffranks/spruce/issues/318).

# Acknowledgments

THanks @ywei2017 for the crash fix, and @VasylTretiakov for the empty doc support!
