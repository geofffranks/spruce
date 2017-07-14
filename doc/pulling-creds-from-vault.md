## Can I use Spruce with Vault?

`spruce` has a `(( vault ))` operator for retreiving credentials from Vault.
Details on using it can be found in the [operator docs][operator-docs].

A common practice is to use `spruce` with the `REDACT=true` environment variable
set, to tell `spruce` to place `REDACTED` in place of the sensitive credentials
during most usage, so credentials aren't accidentally saved in a repository and
published. When the credentials need to be looked up, unset the `REDACT` environment
variable, and spruce merge again to a temporary file, and ensure that it is cleaned
up after being used.

[operator-docs]:        https://github.com/geofffranks/spruce/blob/master/doc/operators.md#-vault-
