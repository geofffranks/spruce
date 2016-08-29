# Modifying arrays
Technically, YAML allows you to create arrays of everything that is supported as a type. However, when dealing with BOSH deployment manifests, you typically encounter either
- an array of simple types such as strings or numbers,
- or an array of maps.
An array of maps follows a common schema: One of the map entries serves as an identifier. This identifier gives the array entry a name and enables it to be referenced it internally.

Simple array:
```yml
list:
- 192.168.1.1
- 192.168.1.2
```

Array of maps:
```yml
list:
- name: Controller
  version: v1
  active: true
- name: Logger
  version: v4
  active: false
```

In this example, the map
```yml
name: Controller
version: v1
active: true
```
is an array entry.

Spruce provides operators to modify a given array:
- `(( append ))`
- `(( prepend ))`
- `(( insert ... ))`
- `(( delete ... ))`
- `(( replace ))`
- `(( merge ))`, and `(( merge ... ))` respectively

Some of these operators can be used multiple times and/or in combination with others. These operators are: `(( append ))`, `(( prepend ))`, `(( insert ... ))`, and `(( delete ... ))`. With the exception of `(( delete ... ))`, the respective operator applies to all following array entries until the end of the array or a new operator. Obviously, the delete operator always stands alone. Array entries that cannot be attached to a preceding operator are called _orphaned_ entries and will result in an error.
```yml
list:
- (( append ))
- "The new first entry"
- "The new second entry"

- (( prepend ))
- "The new second to last entry"
- "The new last entry"
```

To emphasize the new annotation an empty line was added. This is technically not needed, however it increases the readability.

## Operators that work the same for either type of array
- `(( append ))` - The append operator tells Spruce to place the following content after the last existing entry.

- `(( prepend ))` - The prepend operator tells Spruce to place the following content before the first existing entry.

- `(( replace ))` - The replace operator tells Spruce to completely remove the existing array and replace it with the following entries.

- `(( inline ))` - The inline operator might be a bit confusing at first glance. It takes the given entries and puts them on top of the existing ones based on their respective position in the array (index). The first one of the new entries is merged with the first one of the existing array and so on and so forth.
  The inline operator works with both types of arrays, however it is not _recommended_ to be used with arrays of maps, because this greatly reduces the maintainability. This is due to the fact that a change in the order of the existing list can create unexpected results.
  Use this operator if you really know what you are doing and for arrays that do not change very much, that use simple types, or that are relatively small in size.

## Operators that modify arrays of maps (based on identifiers)
There are three well-known key names that serve as identifiers: `name`, `key`, and `id`. The key name `name` is the most common one and heavily used in BOSH deployment manifests.
```yml
jobs:
- name: consul
  instances: 1
  ...
- name: doppler
  instances: 1
  ...
```

This makes is possible to tell Spruce to do modifications based on the name and its position in the array. A modification can be the addition of new configuration fragments (called merging), or the addition of new array entries itself relative to an existing entry.

- `(( merge ))` - The merge operator tells Spruce that the following array entries should be merged with the existing ones based on their name using `name` as the identifier key name. This is the Spruce default behavior for arrays if nothing is specified at all. The operator can therefore be omitted. Again, based on the given example:
  ```yml
  list:
  - name: consul
    instances: 2
    resource_pool: special-pool
  ```
  This will tell Spruce to look for an already existing array entry named `consul`. The existing array entry will then be extended with the provided configuration fragments. If they are already defined, the new values overrules the old one. New key/values will be simply added.
  If no array entry under that name can be found in the existing array, then the new array entry will simply be added to the current list at the end of the existing array.

- `(( merge ... ))` - If you want to merge arrays that do not use `name` as the identifier key name, you have to specify the identifier key name explicitly. For example, you can use `(( merge on id ))` to tell Spruce to merge the new array entries using `id` as the identifier key name.
  ```yml
  components:
  - (( merge on id ))
  - id: ba6f379e-9151-4379-91bf-2f1d15a92902
    active: true
  - id : 3a9c11e3-e45b-4d7d-b020-ca612976c554
    active: false
  ```

- `(( insert ... ))` - The insert operator expects two or three arguments: You tell Spruce whether new entries should be inserted `before` or `after` the insertion point. Furthermore, you specify the insertion point itself by providing the identifier key name and the actual identifiable name.
  The identifier key name can be omitted in which case `name` is used as the default. Based on the given example, if you specify `(( insert after name "consul" ))`, the following entries will be inserted after `consul` and before `doppler`. Since the identifier key name is `name` in our example, you can shorten it to: `(( insert after "consul" ))`.
  The insert operator will not work if either the insertion point cannot be found or one of the new entries is already part the existing array.

- `(( delete ... ))` - Like the insert operator, the delete operator expects arguments to locate the array entry of the existing array that needs to be removed: You provide the identifier key name and the identifiable name.
  Again, the identifier key name can be omitted and defaults to `name`. Based on given example, if you specify `(( delete "consul" ))`, the `consul` entry would be deleted leaving the list with only the `doppler` entry.

## Operators that modify simple arrays (based on index)
Without the possibility to reference an array entry by an identifier, you can only use the index of each array entry as a point of reference for operators that need a specified reference point.

- `(( insert ... ))` - The insert operator supports using an index. The syntax is the same as with names. You can specify `(( insert after 0 ))` to tell Spruce to put the following entries after the first one. This means `(( insert before 0 ))` is equivalent to `(( prepend ))`.

- `(( delete ... ))` - Analog to the insert operator, the delete operator also supports an index as an input argument. For example, `(( delete 0 ))` would remove the first array entry.
