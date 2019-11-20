## Fanning source out to multi-doc YAML files

`spruce fan` will take a source file and multiple target files (multi-doc YAML supported!),
and merge the target documents on top of the source, outputting a giant multi-doc YAML stream.

`spruce fan` takes all the same flags as a spruce merge, and supports merging the same way, with the
exception that only two documents are ever merged together (source, and each target doc independently).

### Examples

source.yml:
```
meta:
  here: is a thing
  these: are things
```

target-1.yml:
```
stuff: (( grab meta.here ))
---
second-doc:
  something nested: (( concat "concatenating " meta.here ))
```

target-2:.yml:
```
---
cats: dogs
---
bats: (( grab meta.these ))
```

Usage:
```
spruce merge --prune meta source.yml target-1.yml target-2.yml
---
stuff: is a thing
---
second-doc:
  something nested: concatenating is a thing
---
cats: dogs
---
bats: are things
```
