# New Features

- The `(( vault ))` operator now allows for multiple arguments.
  It previously required exactly one argument. Now if given multiple
  arguments, they are concatenated together, and used as the path
  to search inside Vault. For example: `(( vault meta.vault_prefix "/key:pass" ))`.
  One-argument `(( vault ))` calls still behave as they used to.
