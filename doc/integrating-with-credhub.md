## Can I use Spruce with CredHub?

Absolutely. In most cases, CredHub credentials are specified in YAML using a syntax
that looks very similar to that of a `spruce` operator. Unfortunately, this will cause
errors with `spruce`, as it thinks it found an unknown operator. To counter-act this,
just use the `((!...))` notation wherever you would use CredHub's `((...))` notation,
like this:

```
---
secret:
  password: ((!from_credhub))
```

This will be ignored in `spruce`, and output `((!from_credhub))` as the value for
`secret.password`. When BOSH reads this in, it will handle dropping the `!` for
you, and all will be well.

## Ingesting templates written for CredHub

If you have an upstream template that you wish to use with `spruce`, you can
run a `sed` on them, to replace the `((...))` notation with `((!...))`, and
 they will be ready to use with `spruce.
