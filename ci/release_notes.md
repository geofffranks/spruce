# New Features

New `(( vault ... ))` operator allows Spruce templates to reach
out to a [Vault][vault] (in a previously authenticated session)
and retrieve secrets like passwords and RSA keys, securely.
See [Spruce, Vault, Concourse & You][blog] for more details.

New `(( cartesian-product ... ))` operator can be used to combine
two or more lists into their normal cartesian product via
concatenation.

Using `(( merge ))` in a non-list context now throws an error, to
assist people converting from spiff.

Spruce releases are now compiled with Go 1.5, producing statically
linked binaries that don't need any shared libraries on the host
system.

# Bug Fixes

Lists are now merged in `(( inject ... ))` calls, so the following
snippet does what you would expect:

```yml
meta:
  api:
    templates:
      - { release: my-release, name: job1 }
      - { release: my-release, name: job3 }

jobs:
  api:
    .: (( inject meta.api_node ))
    templates:
      - { release: my-other-release, name: a-job }
```

The `(( param ... ))` operator now throws an error if an
unoverridden parameter is used as an composite argument to another
operator.  Notably, the following snippet now throws an error:

```yml
meta:
  domain: (( param "You need a system domain" ))

properties:
  endpoint: (( concat "https://api." meta.domain ":8888" ))
```

[vault]: https://vaultproject.io
[blog]:  https://blog.starkandwayne.com/2016/01/11/spruce-vault-concourse-you/
