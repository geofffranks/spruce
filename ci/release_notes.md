# New Features

- `spruce` now supports merging multi-doc YAML files onvia the `fan` subcommand!
  It takes a source yaml file, and fans that out to each doc of any subsequent files (or
  data read from STDIN), combining it all in a giant multi-doc YAML stream.

  Usage:
  ```
  spruce fan my-source.yml multi-doc-file-1.yml ... multi-doc-file-N.yml
  ```

  See https://github.com/geofffranks/spruce/blob/master/doc/fan.md
