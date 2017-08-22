# Windows

Experimental Windows binaries are being shipped alongside spruce now!
If you encouter any strange issues, submit a GH issue.

# Bug Fixes

- `--skip-eval` properly skips `(( param ))` and `(( inject ))` operators now. `prune`
  behavior works as it has in the past (the `(( prune ))` operators are not evaluated,
  but `--prune` arguments are.

- When using the `(( delete ))` array operator, the quotes are no longer required:

  ```
  - (( delete "myObj" ))
  - (( delete myObj ))
  ```
