# Improvements

- Spruce will now ignore things that look like BOSH/Concourse
  variables, so you can pass-through things like `((my-cert))`
  without throwing an error about unknown Spruce operators.
