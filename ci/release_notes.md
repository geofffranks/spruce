# Fixes

- Resolved an issue where the `(( calc ... ))` operator was confounded by certain
  numeric data types, thinking that things like `int64` weren't numbers.

# Acknowledgements

Thanks for the fix @HeavyWombat!
