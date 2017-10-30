# Bug Fixes

- When pulling environment variables in, their values are now
  unmarshaled into appropriate types via YAML. For example, `PORT=123`
  will be used as an integer in the output YAML. This may cause problems
  if a consumer of the output YAML is requiring a string value of the number/boolean
  being displayed. If this is the case, `(( concat $PORT "" ))` can be
  used as a workaround. It seems more likely that there will be consumers
  requiring the value as an integer/boolean type rather the stringified version.

- Resolved an issue where `spruce diff` was not noticing changes in
  type between values. For example, changing from `"123" to `123` 
  will now produce diff output.
