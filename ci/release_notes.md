# New Features

- The `(( load ))` operator now supports loading files based off of references in the YAML
  datastructure. For example:

```
file: loadme.yml

data: (( load file ))
```

# Bug Fixes

- The `(( load ))` operator correctly handles absolute file paths now. Previously, it
  would detect absolute file paths as URLs and fail to retreive them.

# Acknowledgements

Thanks @dennisjbell for the bug fix/feature!
