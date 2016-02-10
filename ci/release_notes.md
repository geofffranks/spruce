# New Features

- `spruce json` is a new subcommand. It allows you to pipe yaml through spruce
and get json-compatible output.

# Bug Fixes
- Fixed the `(( vault ))` operator to honor tokens generated from `vault auth`
- Fixed the `(( vault ))` operator to handle HTTP redirects from the Vault API

# Miscellaneous

We've updated the CI pipeline to be more awesome. It now generates + uploads
binaries directly to the release as artifacts, without the silly gzip + version
number embedding that used to occur. If you have anything that is automatically
downloading `spruce` from the latest release, it is likely that you will need to
update your workflow. **(THIS IS A BREAKING CHANGE)**
