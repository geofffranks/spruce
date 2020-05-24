## Pointer syntax

- pointers always start at the root of the document

- strings typically refer to hash keys (ex: `/key1`)
  - strings ending with `?` refer to hash keys that may or may not exist
    - "optionality" carries over to the items to the right

- integers refer to array indices (ex: `/0`, `/-1`)

- `-` refers to an imaginary index after last array index (ex: `/-`)

- `key=val` notation matches hashes within an array (ex: `/key=val`)
  - values ending with `?` refer to array items that may or may not exist

- array index selection could be affected via `:prev` and `:next`

- array insertion could be affected via `:before` and `:after`

See pointer test examples in [patch/pointer_test.go](../patch/pointer_test.go).

## Operations

Following example is used to demonstrate operations below:

```yaml
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
```

There are two available operations: `replace` and `remove`.

### Hash

```yaml
- type: replace
  path: /key
  value: 10
```

- sets `key` to `10`

```yaml
- type: replace
  path: /key_not_there
  value: 10
```

- errors because `key_not_there` is expected (does not have `?`)

```yaml
- type: replace
  path: /new_key?
  value: 10
```

- creates `new_key` because it ends with `?` and sets it to `10`

```yaml
- type: replace
  path: /key2/nested/super_nested
  value: 10
```

- requires that `key2` and `nested` hashes exist
- sets `super_nested` to `10`

```yaml
- type: replace
  path: /key2/nested?/another_nested/super_nested
  value: 10
```

- requires that `key2` hash exists
- allows `nested`, `another_nested` and `super_nested` not to exist because `?` carries over to nested keys
- creates `another_nested` and `super_nested` before setting `super_nested` to `10`, resulting in:

  ```yaml
  ...
  key2:
    nested:
      another_nested:
        super_nested: 10
      super_nested: 2
    other: 3
  ```

### Array

```yaml
- type: replace
  path: /array/0
  value: 10
```

- requires `array` to exist and be an array
- replaces 0th item in `array` array with `10`

```yaml
- type: replace
  path: /array/-
  value: 10
```

- requires `array` to exist and be an array
- appends `10` to the end of `array`

```yaml
- type: replace
  path: /array2?/-
  value: 10
```

- creates `array2` array since it does not exist
- appends `10` to the end of `array2`

```yaml
- type: replace
  path: /array/1:prev
  value: 10
```

- requires `array` to exist and be an array
- replaces 0th item in `array` array with `10`

```yaml
- type: replace
  path: /array/0:next
  value: 10
```

- requires `array` to exist and be an array
- replaces 1st item (starting at 0) in `array` array with `10`

```yaml
- type: replace
  path: /array/0:after
  value: 10
```

- requires `array` to exist and be an array
- inserts `10` after 0th item in `array` array

```yaml
- type: replace
  path: /array/0:before
  value: 10
```

- requires `array` to exist and be an array
- inserts `10` before 0th item at the beginning of `array` array

### Arrays of hashes

```yaml
- type: replace
  path: /items/name=item7/count
  value: 10
```

- finds array item with matching key `name` with value `item7`
- adds `count` key as a sibling of name, resulting in:

	```yaml
	...
	items:
	- name: item7
	  count: 10
	- name: item8
	```

```yaml
- type: replace
  path: /items/name=item8/count
  value: 10
```

- errors because there are two values that have `item8` as their `name`

```yaml
- type: replace
  path: /items/name=item9?/count
  value: 10
```

- appends array item with matching key `name` with value `item9` because values ends with `?` and item does not exist
- creates `count` and sets it to `10` within created array item, resulting in:

  ```yaml
  ...
  items:
  - name: item7
  - name: item8
  - name: item8
  - name: item9
    count: 10
  ```

See full example in [patch/integration_test.go](../patch/integration_test.go).
