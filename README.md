```
          *          .---. ,---.  ,---.  .-. .-.  ,--,  ,---.         *
         /.\        ( .-._)| .-.\ | .-.\ | | | |.' .')  | .-'        /.\
        /..'\      (_) \   | |-' )| `-'/ | | | ||  |(_) | `-.       /..'\
        /'.'\      _  \ \  | |--' |   (  | | | |\  \    | .-'       /'.'\
       /.''.'\    ( `-'  ) | |    | |\ \ | `-')| \  `-. |  `--.    /.''.'\
       /.'.'.\     `----'  /(     |_| \)\`---(_)  \____\/( __.'    /.'.'.\
"'""""/'.''.'.\""'"'""""""(__)""""""""(__)"""""""""""""(__)""'""""/'.''.'.\""'"'"
      ^^^[_]^^^                                                   ^^^[_]^^^
```

[![Build Status](https://img.shields.io/travis/geofffranks/spruce.svg)](https://travis-ci.org/geofffranks/spruce)

## Introducing Spruce

`spruce` is a domain-specific YAML merging tool, for generating [BOSH](http://bosh.io) manifests.

It was written with the goal of being the most intuitive solution for merging BOSH templates.
As such, it pulls in a few semantics that may seem familiar to those used to merging with [the other merging tool](https://github.com/cloudfoundry-incubator/spiff),
but there are a few key differences.

## Installation

You can download a [prebuilt binaries for 64-bit Linux, or Mac OS X](https://github.com/geofffranks/spruce/releases/),
or you can install via `go get` (provided you have installed go):

```
go get github.com/geofffranks/spruce
```

## Merging Rules

Merging in `spruce` is designed to be pretty intuitive. Files to merge are listed
in-order on the command line. The first file serves as the base to the file structure,
and subsequent files are merged on top, adding when keys are new, replacing when keys
exist. This differs slightly in mentality from spiff, but hopefully the results are
more predictable.

### A word on 'meta'

`meta` was a convention used quite often in templates merged with spiff. This convention
is not necessary with spruce. If you want to merge two hashes together, simply include
the new keys in the file merged on top of the original.

### What about arrays?

Arrays can be merged in three ways - prepending data, appending data, and completely replacing data.

- To append data to an existing array, ensure that the first element in the new array is <br>

  ```yml
  - (( append ))
  ```

- To prepend the data to an existing array, ensure that the first element in the new array is <br>

  ```yml
  - (( prepend ))
  ```

- To merge the two arrays together (each index of the new array will be merged into the original, additionals appended),
  ensure that the first element in the new array is <br>

  ```yml
  - (( inline ))
  ```

- To merge two arrays of maps together (using a specific key for identifying like objects), ensure that the first element
  in the new array is either <br>

  ```yml
  - (( merge ))
  ```

  <br> or <br>

  ```yml
  - (( merge on <key> ))
  ```

<br> The first merges using `name` as the key to determine
  like objects in the array elements. The second is used to customize which key to use. See [Merging Arrays of Maps](#mapmerge)
  for an example.

- To completely replace the array, don't do anything special - just make the new array what you want it to be!

### Cleaning Up After Yourself

To prune a map key from the final output<br>

  ```yml
  useless: (( prune ))
  ```

### Referencing Other Data

Need to reference existing data in your datastructure? No problem! `spruce` will wait until
all the data is merged together before dereferencing anything, but to handle this, you can
use the `(( grab <thing> )) syntax:

```yml
data:
  color: blue

pen:
  color: (( grab data.color ))
```

###Hmm.. How about auto-calculating resource pool sizes, and static IPs?

That's a great question, and soon, spruce will support that!

## How About an Example?

### Basic Example

Here's a pretty broad example, that should cover all the functionality of spruce, to be used as a reference.

If I start with this data:

```yml
top:
  orig_key: This is a string attached to a key
  number: 50
  array1:
  - first element
  - second element
  - third element
  map:
    key1: v1
    key2: v2
    key3:
      subkey1: vv1
      subkey2: vv2
      subkey3:
      - nested element 1
      - nested element 2

  1: 430.0
  2: this starts as a string
  array2:
  - 1
  - 2
  - 3
  - 4
  inline_array_merge:
  - will be overwritten
  - this: will
    be: merged
```

And want to merge in this:

```yml
top:
  new_key: this is added
  orig_key: this is replaced
  map:
    key4: added key
    key1: replaced key
    key2: ~
    key3:
      subkey3:
      - (( append ))
      - nested element 3
  array1:
  - (( prepend ))
  - prepend this
  array2:
  - over
  - ridden
  - array
  1: You can change types too
  2:
    even: drastically
    to:   from scalars to maps/lists
  inline_array_merge:
  - (( inline ))
  - this has been overwritten
  - be: overwritten
    merging: success!
othertop: you can add new top level keys too
```

I would use `spruce` like this:

```yml
$ spruce merge assets/examples/example.yml assets/examples/example2.yml
othertop: you can add new top level keys too
top:
  1: 430
  2: this starts as a string
  "1": You can change types too
  "2":
    even: drastically
    to: from scalars to maps/lists
  array1:
  - prepend this
  - first element
  - second element
  - third element
  array2:
  - over
  - ridden
  - array
  inline_array_merge:
  - this has been overwritten
  - this: will
    be: overwritten
    merge: success!
  map:
    key1: replaced key
    key2: null
    key3:
      subkey1: vv1
      subkey2: vv2
      subkey3:
      - nested element 1
      - nested element 2
      - nested element 3
    key4: added key
  new_key: this is added
  number: 50
  orig_key: this is replaced
```

### Map Replacements

One of [spiff's](https://github.com/cloudfoundry-incubator/spiff) quirks was that it quite easily allowed you to completely replace an
entire map, with new data (rather than merging by default). That result is still
possible with `spruce`, but it takes a little bit more work, since the primary
use case is to merge two maps together:

We start with this yaml:

```yml
$ cat original.yml
untouched:
  map: stays
  the: same
map_to_replace:
  has: upstream
  data: that
  we: do
  not: want
```

Next, create a YAML file to clear out the map:
```yml
$ cat clear.yml
map_to_replace: ~
```

Now, create a YAML file to insert the data you want in the end:
```yml
$ cat new.yml
map_to_replace:
  my: special
  data: here
```

And finally, merge it all together:

```yml
$ spruce merge original.yml clear.yml new.yaml
map_to_replace:
  my: special
  data: here
untouched:
  map: stays
  the: same
```

*NOTE:* due to map key randomizations, the actual order of the above output will vary.

### Key Removal

How about deleting keys outright?

```yml
---
# original.yml
meta:
  thing: &thing
    foo: 1
    bar: 2

things:
  - <<: *thing
    name: first-thing
  - <<: *thing
    name: second-thing
```

The `meta` key is only useful for holding the `*thing` referent,
so we'd really rather not see it in the final output:

```yml
---
# prune.yml
meta: (( prune ))
```

```yml
$ spruce merge original.yml prune.yml
things:
  - name: first-thing
    foo: 1
    bar: 2
  - name: second-thing
    foo: 1
    bar: 2

```


###<a name="mapmerge"></a>Merging Arrays of Maps

Let's say you have a list of maps that you would like to merge into another list of maps, while preserving
as much data as possible.

Given this `original.yml`:
```yml
jobs:
- name: concatenator_z1
  network: generic1
  resource_pool: small
  properties:
    spruce: is cool
- name: oldjob_z1
  network: generic1
  resource_pool: small
  properties:
    this: will show up in the end
```

And this `new.yml`:
```yml
jobs:
- name: newjob_z1
  network: generic1
  resource_pool: small
  properties:
    this: is a job defined solely in new.yml
- name: concatenator_z1
  properties:
    this: is a new property added to an existing job
```

You would get this when merged:
```yml
$ spruce merge original.yml new.yml
jobs:
- name: concatenator_z1
  network: generic1
  properties:
    spruce: is cool
    this: is a new property added to an existing job
  resource_pool: small
- name: oldjob_z1
  network: generic1
  properties:
    this: will show up in the end
  resource_pool: small
- name: newjob_z1
  network: generic1
  properties:
    this: is a job defined solely in new.yml
  resource_pool: small
```

Pretty sweet, huh?

## Author

Written By Geoff Franks, inspired by [spiff](https://github.com/cloudfoundry-incubator/spiff)

## License

Licensed under [the Apache License v2.0](https://github.com/geofffranks/spruce/raw/master/LICENSE)
