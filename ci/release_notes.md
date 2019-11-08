# Improvements

- `spruce` will automatically unmarshal the contents of environment variables as yaml,
  to turn boolean/numerics into native types, rather than strings. It also is possible
  to use this to unmarshal json objects + arrays into the yaml datastructure. If
  pulling in the data without unmarshaling is desired, you can now specify the 
  `SPRUCE_NO_PARSE_ENV_VARS_AS_YAML` environment variable to `true`, to disable that
  behavior. Use this with caution, as it will the feature globally, so any environment
  variables needed as boolean or numeric types will instead become strings.
