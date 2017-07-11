# New Features

- `spruce` now supports [go-patch](https://github.com/cppforlife/go-patch) files
  in its `merge` phase via the `--go-patch` flag. This means you can interleave
  traditional spruce/yaml files with go-patch files. The go-patch files will be
  merged into the document, and can even be used to insert spruce operators to
  be evaluated later.

  For example: `spruce merge --go-patch base.yml patch.yml more-spruce.yml` will apply
  the go-patch file `patch.yml` on top of `base.yml`, and then merge in `more-spruce.yml`
  via traditional spruce merging logic. Once merging is complete, the other 
  evaluation phases take place, executing operators, requiring params, etc.

- Added the `(( defer ))` operator. This allows you to specify
  an operation in your yaml that you wish to defer until a later sprucing.
  This can be useful for using spruce to generate more spruce templates,
  or to handle content that requires `(( ... ))` syntax, like CredHub.

  For example:

  ```
  still_a_grab_op: (( defer grab myval ))
  myval: 1
  ```

  Would yield:

  ```
  still_a_grab_op: (( grab myval ))
  myval: 1
  ```

  when merged a single time. Running it through spruce again would result
  in the grab being evaluated.

- Not technically a new feature, but newly documented as a feature is the
  `((! ... ))` syntax. Spruce will ignore this completely, and not attempt
  to evaluate any operations on it. Unlike defer, the output will be identical
  to the input (the exclamation point is kept in the output).

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

- Fixes #153 and #169. The result of the cartesian-product operator
  now behaves as it should in join/concat/inject and other operators.

- Integers above a 64-bit unsigned quantity are now supported in operations.
  They are automatically converted to scientific notation, and treated as floats.

# Acknowledgements

Thanks to @thomasmmitchell, @jhunt, and @poblin-orange for their help on all the features
and fixes in this release!
