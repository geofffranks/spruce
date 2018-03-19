# Removals

- The long-deprecated `--concourse` flag has been removed from `spruce`. Please use
  either the `(( grab ))` operator, or Concourse's `((variable))` syntax for pulling
  in the credentials.
