# Improvements

- If you need to change the way maps are merged throughout an entire file,
  the `DEFAULT_ARRAY_MERGE_KEY` environment variable can be set, to override
  the default value of 'name'. This makes it easy to not need to put `(( merge on identifier ))`
  everywhere in your YAML, when `name` is not the desired key.

# Bug Fixes

- Fixed an issue when grabbing multi-line string values from environment variables with
  `(( grab $MULTI_LINE_STRING_VAR ))`

# Acknowledgements

Thanks for the bugfix @sorenh!
