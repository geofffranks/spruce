# New Features

Added `(( file ... ))` operator to include file contents as text blocks to
allow file to be retained in it natural form.

* Can use literal string or reference as argument
* Relative to path specified in `SPRUCE_FILE_BASE_PATH` environment variable,
  or whatever directory spruce is run in if not set.
* Absolute paths can also be set -- prefix with / to specify.
