## Can I use Spruce with Vault?

`spruce` has a `(( vault ))` operator for retreiving credentials from Vault.
Details on using it can be found in the [operator docs][operator-docs].

A common practice is to use `spruce` with the `REDACT=true` environment variable
set, to tell `spruce` to place `REDACTED` in place of the sensitive credentials
during most usage, so credentials aren't accidentally saved in a repository and
published. When the credentials need to be looked up, unset the `REDACT` environment
variable, and spruce merge again to a temporary file, and ensure that it is cleaned
up after being used.

By default `spruce` use [Vault KV v1](https://www.vaultproject.io/api/secret/kv/kv-v1.html).
To use newer [Vault KV v2](https://www.vaultproject.io/api/secret/kv/kv-v2.html) set environment
variable `VAULT_VERSION` to `2` (default: `1`). Currently it's not possible to access both
types of KV Engines at the same time.

Here's an example:

```
$ cat <<EOF base.yml

credentials:
- username: (( vault "secret/my/credentials/admin:username" ))
  password: (( vault "secret/my/credentials/admin:password" ))
EOF

$ REDACT=yes spruce merge base.yml
credentials:
- password: REDACTED
  username: REDACTED

$ spruce merge base.yml
credentials:
- password: thisPasswordWasPulledFromVault
  username: adminUserNamePulledFromVault
```

In the above example, there was a path in the Vault `secret` backend of
`secret/my/credentials/admin`. That path contained two keys `username`,
and `password`, set to `adminUserNamePulledFromVault`, and `thisPasswordWasPulledFromVault`.

[operator-docs]:        https://github.com/geofffranks/spruce/blob/master/doc/operators.md#-vault-
