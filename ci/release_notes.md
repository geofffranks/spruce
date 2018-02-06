#Bug Fixes

- `spruce diff` output used to be colorized based on whether or not STDERR was 
  sent to a terminal. This led to issues piping or redirecting the actual diff output,
  so now the decision is based on what type of device STDOUT is. All other `spruce`
  subcommands are unaffected, and continue to colorize based on STDERR's device type.

  Thanks @giner for pointing this out!
