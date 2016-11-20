## New Features

- Experimental support for a `spruce diff` command that generates
  semantic differential reports highlighting the differences
  between two YAML files.  The differences are semantic, meaning
  that they do not care about whitespace, flow-vs-block, or map
  key ordering.  The produced diffs are stable, so that if you run
  `spruce diff` twice, without changing input files, you get the
  same output -- this is surprisingly helpful when trying to
  reconcile to large YAML documents.
