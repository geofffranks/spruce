## Improvements

- Environment variables can now be lowercase, allowing constructs
  like `(( grab $http_proxy ))`.  Fixes #177

- The `(( vault ))` operator handles paths with colons in them in a way
  that matches how the `safe` binary operates. Both treat the text after the
  **last** colon to be the key, and everything prior as the key. For example:

  ```
  (( vault "secret/path:to:my:value:key" )) # finds the `key` value inside `secret/path:to:my:value`
  ```

- Forked our yaml parsers + applied a patch to resove https://github.com/go-yaml/yaml/issues/75
