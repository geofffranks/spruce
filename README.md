## Introducing Spruce

`spruce` is a domain-specific YAML merging tool, for generating [BOSH](http://bosh.io) manifests.

It was written with the goal of being the most intuitive solution for merging BOSH templates.
As such, it pulls in a few semantics that may seem familiar to those used to merging with [the other merging tool](https://github.com/cloudfoundry-incubator/spiff),
but there are a few key differences.

## Installation

You can download a [prebuilt binaries for 64-bit linux, or Mac OS X](https://github.com/geofffranks/spruce/releases/),
or you can install via `go get` (provided you have installed go):

```
go get github.com/geofffranks/spruce
```

## Merging Rules

Merging in `spruce` is designed to be pretty intuiive. Files to merge are listed
in-order on the command line. The first file serves as the base to the file structure,
and subesquent files are merged on top, adding when keys are new, replacing when keys
exist. This differs slightly in mentality from spiff, but hopefully the results are
more predictable.

### A word on 'meta'

`meta` was a convention used quite often in templates merged with spiff. This convention
is not necessary with spruce. If you want to merge two hashes together, simply include
the new keys in the file merged on top of the original.

### What about arrays?

Arrays can be merged in three ways - prepending data, appending data, and completely replacing data.

- To append data to an existing array, ensure that the first element in the new array is `(( merge ))`
- To prepend the data to an existing array, ensure that the *last* element in the new array is `(( merge ))`
- To completely replace the array, don't do anything special - just make the new array what you want it to be!

### Hmm.. How about auto-calculating resource pool sizes, and static IPs?

That's a great question, and soon, spruce will support that!

## How About an Example?

Here's a pretty broad example, that should cover all the functionality of spruce, to be used as a reference.

If I start with this data:

```
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
```

And want to merge in this:

```
top:
  new_key: this is added
  orig_key: this is replaced
  map:
    key4: added key
    key1: replaced key
    key2: ~
    key3:
      subkey3:
      - (( merge ))
      - nested element 3
  array1:
  - prepend this
  - (( merge ))
  array2:
  - over
  - ridden
  - array
  1: You can change types too
  2:
    even: drastically
    to:   from scalars to maps/lists
othertop: you can add new top level keys too
```

I would use `spruce` like this:

```
$ ./spruce assets/examples/example.yml assets/examples/example2.yml
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

$
```

## Author

Written By Geoff Franks, inspired by [spiff](https://github.com/cloudfoundry-incubator/spiff)

## License

Licensed under [the Apache License v2.0](https://github.com/geofffranks/spruce/raw/master/LICENSE)
