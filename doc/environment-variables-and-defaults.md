## Can I specify a default for an operation?

You certainly can, via the magic of `||`:

```
value: (( grab original.value || "default-value" ))
```

If spruce cannot find reference `original.value` in the root document,
it fails over to the literal `default-value`. The logical-or can be used
at any point a `spruce` operator takes an argument. If the argument on the
left side of the `||` cannot be resolved, the argument on the right side is
used. 

Right-side arguments do not need to be literal values. They can also be references:

```
value: (( grab original.value || failover.value ))
```

In this case, if `original.value` is missing, `failover.value` will be tried.
If it is also missing, `spruce` will throw an error indicating that `failover.value`
could not be found.

It is also possible to chain `||` multiple times:

```
value: (( grab original.value || failover.value || "default-value" ))
```

## Can I use the values of environment variables in an operation?

Again, the answer is yes! Here's how:

```
envVar: (( grab $MY_ENVIRONMENT_VARIABLE ))
```

Spruce will find the value of the requested environment variable (erroring if it is not set),
and pass it in to the operator as a literal value. Environment variables cannot be used to dynamically
reference parts of the root document in an operator. Their values are currently always literals when
given to the operator. If you have a use case that needs this, stick in an [issue][issues], and
we'll see if we can work in support for this.

[issues]: https://github.com/geofffranks/spruce/issues/new
