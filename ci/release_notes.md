# New Features

New TRACE mode for debugging (activate via `--trace`) that will
provide even more verbose debugging output.  Anything printed
under `--debug` is not intended to help template authors
troubleshoot merge, override, and operation problems with their
templates.  `--trace`-level debugging is more for assisting spruce
developers in reproducing buggy behavior, or for those of you who
are *reallllly* interested in spruce internals!

New `(( inject ... ))` operator for duplicating parts of the YAML
structure elsewhere.  These injected sub-structures can then be
overridden at the local level.

Spruce internals got a major overhaul, including a new
merge-eval-check phased approach to template rendering.  During
each phase, spruce performs static data-flow analysis to determine
what operations need to be called before what other operations.
This leads to a better end-user experience, and fixes several
classes of bugs related to implicit dependencies and call
ordering.

Operator arguments are now evaluated according to a simple set of
expression rules, allowing alternatives to be specified like so:

```
(( grab first.option || second.option || nil ))
```


# Bug Fixes

Miscellaneous code cleanups, reformatting and documentation typo and broken
link fixes went into this release.

Specifically fixed some breakage related to numeric keys, as demonstrated in
the examples in the README file.  To catch these in the future, the README
examples are now integrated into the unit tests to verify the documentation.


# Acknowledgements

Many thanks to [James Hunt](https://github.com/filefrog) for `inject` operator, `TRACE` mode,
and internals overhaul!
