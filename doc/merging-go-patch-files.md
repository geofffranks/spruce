## Can `spruce` be used in conjunction with [go-patch][gopatch]?

Yes! As of v1.11.0, `spruce merge --go-patch` was added. When merging files, it will detect if a
file is in the [go-patch][gopatch] format, parse the operations defined inside it, and execute
those on the root document. This means you can mix and match regular YAML, with `spruce` operators
with [go-patch][gopatch] files, allowing you to use upstream templates and BOSH ops-files, with
custom `spruce`-based templates on top. Since this all occurs during the **Merge Phase**, it means
you can even use go-patch to insert `spruce` operators into the datastructure!

Here's an example:

```
# Base yaml
$ cat <<EOF > base.yml
key: 1

key2:
  nested:
    super_nested: 2
  other: 3

array: [4,5,6]

items:
- name: item7
- name: item8
- name: item8
EOF

# the go-patch file
$ cat ,<EOF > go-patch.yml
- type: replace
  path: /key
  value: 10

- type: replace
  path: /new_key?
  value: 10

- type: replace
  path: /key2/nested/super_nested
  value: 10

- type: replace
  path: /key2/nested?/another_nested/super_nested
  value: 10

- type: replace
  path: /array/0
  value: 10

- type: replace
  path: /spruce_array_grab?
  value: (( grab items ))
EOF

#  another spruce file
$ cat <<EOF > final.yml
more_stuff: is here

items:
- (( prepend ))
- add spruce stuff in the beginning of the array
EOF

$ spruce merge --go-patch base.yml go-patch.yml final.yml
array:
- 10
- 5
- 6
items:
- add spruce stuff in the beginning of the array
- name: item7
- name: item8
- name: item8
key: 10
key2:
  nested:
    another_nested:
      super_nested: 10
    super_nested: 10
  other: 3
more_stuff: is here
new_key: 10
spruce_array_grab:
- add spruce stuff in the beginning of the array
- name: item7
- name: item8
- name: item8
```

[gopatch]: https://github.com/cppforlife/go-patch
