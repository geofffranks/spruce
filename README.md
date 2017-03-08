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

Questions? Pop in our [slack channel](https://cloudfoundry.slack.com/messages/spruce/)!

## Introducing Spruce

`spruce` is a general purpose YAML & JSON merging tool.

It was originally written with the goal of being the most intuitive solution for merging BOSH templates.
As such, it pulls in a few semantics that may seem familiar to those used to merging with [the other merging tool](https://github.com/cloudfoundry-incubator/spiff),
but there are a few key differences.

## Installation

Spruce is now available via Homebrew, just `brew tap starkandwayne/cf; brew install spruce`

Alternatively,  you can download a [prebuilt binaries for 64-bit Linux, or Mac OS X](https://github.com/geofffranks/spruce/releases/),
or you can install via `go get` (provided you have installed go):

```
go get github.com/geofffranks/spruce/...
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

Arrays can be modified in multiple ways: prepending data, appending data, inserting data, merging data onto the existing data, or completely replacing data.

- To append data to an existing array, ensure that the first element in the new array is <br>

  ```yml
  - (( append ))
  ```

- To prepend the data to an existing array, ensure that the first element in the new array is <br>

  ```yml
  - (( prepend ))
  ```

- To insert new elements either after or before a specific position of an existing array, you can use `insert` with
  a hint to the respective insertion point in the target list <br>

  ```yml
  jobs:
  - name: consul
    instances: 1
  - name: nats
    instances: 2
  - name: ccdb
    instances: 2
  - name: uaadb
    instances: 2
  - name: dea
    instances: 8
  - name: api
    instances: 2
  ```

  ```yml
  jobs:
  - (( insert after "dea" ))
  - name: dea_v2
    instances: 2
  ```

  or <br>

  ```yml
  - (( insert after <key> "<name>" ))
  - <key>: new-kid-on-the-block
  ```
  <br> The first `insert` is using `name` as the default key to determine the target position in the target array.
  The second is used to customize which key to use. In any case, instead of `after`, you can also use `before`. This will
  prepend the entries (relative to the specified insertion point).

- Similar to the `insert` operation, you can also use a `(( delete ... ))` operation multiple times in a list. The `delete`
  will remove a map from the list <br>

  ```yml
  jobs:
  - (( delete "dea" ))
  - (( delete "api" ))
  ```
  or <br>

  ```yml
  - (( delete <key> "<name>" ))
  ```
  <br> The array modification operations `(( append ))`, `(( prepend ))`, `(( delete ... ))`, and `(( insert ... ))` can be
  used multiple times in one list. Entries that follow `(( append ))`, `(( prepend ))`, and `(( insert ... ))` belong to the
  same operation. This however does not apply to `(( delete ... ))`, which always stands alone.

- To replace the first array with the second,
  ensure that the first element in the new array is <br>

  ```yml
  - (( replace ))
  ```

- To merge two arrays by way of their index, just make the first
  element <br>

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
  like objects in the array elements. The second is used to customize which key to use.

- If you don't specify a specific merge strategy, the array will
  be merged automatically; using keys if they exist (i.e. `((
  merge ))`, and array indices otherwise (`(( inline ))`).

Do you want to learn about array modifications in more detail? See [modifying arrays](/doc/arrays.md) for examples and explanations.

### Cleaning Up After Yourself

To prune a map key from the final output, you can either use the `--prune` flag:<br>

```
spruce merge --prune key.1.to.prune --prune key.2.to.prune file1.yml file2.yml
```

or you can use the `(( prune ))` operator:

```
key_to_prune: (( prune ))
```

If you actually want to prune everything but one or two paths from your YAML, you can use the `--cherry-pick` flag to only select what you want to have in the end:<br>

```
spruce merge --cherry-pick jobs --cherry-pick properties file1.yml file2.yml
```

The `--cherry-pick` flag can be used in combination with the `--prune` flag as long as you do not prune the exact path you are about to cherry-pick.

### Referencing Other Data

Need to reference existing data in your datastructure? No problem! `spruce` will wait until
all the data is merged together before dereferencing anything, but to handle this, you can
use the `(( grab <thing> ))` syntax:

```yml
data:
  color: blue

pen:
  color: (( grab data.color ))
```

You can even reference multiple values at once, getting back an array of their data,
for things like getting all IPs of multi-AZ jobs in a BOSH manifest, just do it like so:

```
(( grab jobs.myJob_z1.networks.myNet1.static_ips jobs.myJob_z2.networks.myNet2.static_ips ))
```

You can also provide alternatives to your `grab` operation, by
using the `||` (or) operator:

```
key:      (( grab site.key || nil ))
domain:   (( grab global.domain || "example.com" ))
protocol: (( grab site.protocol || global.protocol || "http" ))
```

In these examples, if the referenced key does not exist, the next
reference is attempted, or the literal value (nil, numbers or
strings) is used.  Spruce recognizes the following keywords and
uses the appropriate literal value:

  - `nil`, `null` and `~` map to the YAML null value
  - `true` is the YAML boolean value for truth
  - `false` is the YAML boolean value for non-truth

Other types of literals include double-quoted strings (with
embedded double quotes escaped with a single backslash - \\),
integer literals (a string of digits) and floating point literals
(a string of digits, a period, and another string of digits).
Scientific notation is not currently supported.

### Accessing the Environment

Want to pull in secret credentials from your environment?  No
problem!

```yml
secrets:
  aws:
    access_key: (( grab $AWS_ACCESS_KEY ))
    secret_key: (( grab $AWS_SECRET_KEY ))
```

`spruce` will try to pull the named environment variables value
from the environment, and fail if the value is not set, or is
empty.  You can use the `||` syntax to provide defaults, á la:

```yml
meta:
  environment: (( grab $ENV_NAME || "default-env" ))
```

### Hmm.. How about merging Spruce stubs without evaluating the commands? 

In some cases you may want to merge 2 or more Spruce files into one so that you can apply their logic with a single file. In other to do that, you can make use of the option `--no-eval` in merge. 

Check out the [no-eval example](#no-eval)

### Hmm.. How about auto-calculating static IPs for a BOSH manifest?

`spruce` supports that too! Just use the same `(( static_ips(x, y, z) ))` syntax
that you're used to with [spiff](https://github.com/cloudfoundry-incubator/spiff),
to specify the offsets in the static IP range for a job's network.

Behind the scenes, there are a couple behavior improvements upon spiff. First,
since all the merging is done first, then post-processing, there's no need
to worry about getting the instances + networks defined before `(( static_ips() ))`
is merged in. Second, the error messaging output should be a lot better to aid in
tracking down why `static_ips()` calls fail.

Check out the [static_ips() example](#static-ips)

### Hmm.. How about calculating some simple mathematical expressions?

This is also possible with `spruce`. You can use the `(( calc ... ))` operator to
specify a mathematical expression. Inside these expressions you can reference
other values using the same path syntax like `grab`. The mathematical expression
needs to be put in quotes since it can contain parenthesis and white spaces
between your operators, references, or numbers.

```yml
meta:
  offset: 2

jobs:
- name: one
  instances: (( calc "meta.offset + 4" ))
```

If you have more sophisticated calculations in mind, you can use these built-in
functions inside your expressions: `max`, `min`, `mod`, `pow`, `sqrt`, `floor`, and `ceil`.

Maybe you want to calculate instance counts based on given target value.

```yml
meta:
  target_mem: 144

jobs:
- name: big_ones
  instances: (( calc "floor(meta.target_mem / 32)" ))

- name: small_ones
  instances: (( calc "floor((meta.target_mem - jobs.big_ones.instances * 32) / 16)" ))
```

### But I Want To Make Strings!!

Yeah, `spruce` can do that!

```yml
env: production
cluster:
  name: mjolnir
ident: (( concat cluster.name "//" env ))
```

Which will give you an `ident:` key of "mjolnir/production"

But what if I have a list of strings that I want as a single line? Like a users list, authorities, or similar.
Do I have to `concat` that piece by piece? No, you can use `join` to concatenate a list into one entry.

```yml
meta:
  authorities:
  - password.write
  - clients.write
  - clients.read
  - scim.write

properties:
  uaa:
    clients:
      admin:
        authorities: (( join "," meta.authorities ))
```

This will give you a concatenated list for `authorities`:
```yml
properties:
  uaa:
    clients:
      admin:
        authorities: password.write,clients.write,clients.read,scim.write
```

## How About Some Examples?

<a name="ex-basic"></a>
### Basic Example

Here's a pretty broad example, that should cover all the functionality of spruce, to be used as a reference.

If I start with this data:

```yml
# examples/basic/main.yml
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
# examples/basic/merge.yml
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
$ spruce merge main.yml merge.yml
othertop: you can add new top level keys too
top:
  1: You can change types too
  2:
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
  - 4
  inline_array_merge:
  - this has been overwritten
  - be: overwritten
    merging: success!
    this: will
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



<a name="ex-map-replacement"></a>
### Map Replacement

One of [spiff's](https://github.com/cloudfoundry-incubator/spiff) quirks was that it quite easily allowed you to completely replace an
entire map, with new data (rather than merging by default). That result is still
possible with `spruce`, but it takes a little bit more work, since the primary
use case is to merge two maps together:

We start with this yaml:

```yml
# examples/map-replacement/original.yml
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
# examples/map-replacement/delete.yml
map_to_replace: ~
```

Now, create a YAML file to insert the data you want in the end:

```yml
# examples/map-replacement/insert.yml
map_to_replace:
  my: special
  data: here
```

And finally, merge it all together:

```yml
$ spruce merge original.yml delete.yml insert.yml
map_to_replace:
  my: special
  data: here
untouched:
  map: stays
  the: same
```



<a name-"ex-key-removal"></a>
### Key Removal

How about deleting keys outright? Use the --prune flag to the merge command:

```yml
# examples/key-removal/original.yml
deleteme:
  thing:
    foo: 1
    bar: 2
```

```yml
# examples/key-removal/things.yml
things:
- name: first-thing
  foo: (( grab deleteme.thing.foo ))
- name: second-thing
  bar: (( grab deleteme.thing.bar ))
```

```
$ spruce merge --prune deleteme original.yml things.yml
```

The `deleteme` key is only useful for holding a temporary value,
so we'd really rather not see it in the final output. `--prune` drops it.



<a name="ex-list-of-maps"></a>
### Lists of Maps

Let's say you have a list of maps that you would like to merge into another list of maps, while preserving
as much data as possible.

Given this `original.yml`:

```yml
# examples/list-of-maps/original.yml
jobs:
- name: concatenator_z1
  instances: 5
  resource_pool: small
  properties:
    spruce: is cool
- name: oldjob_z1
  instances: 4
  resource_pool: small
  properties:
    this: will show up in the end
```

And this `new.yml`:

```yml
# examples/list-of-maps/new.yml
jobs:
- name: newjob_z1
  instances: 3
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
- instances: 5
  name: concatenator_z1
  properties:
    spruce: is cool
    this: is a new property added to an existing job
  resource_pool: small
- instances: 4
  name: oldjob_z1
  properties:
    this: will show up in the end
  resource_pool: small
- instances: 3
  name: newjob_z1
  properties:
    this: is a job defined solely in new.yml
  resource_pool: small
```

Pretty sweet, huh?



<a name="ex-static-ips"></a>
### Static IPs

Lets define our `jobs.yml`:

```yml
# examples/static-ips/jobs.yml
jobs:
- name: staticIP_z1
  instances: 3
  networks:
  - name: net1
    static_ips: (( static_ips(0, 2, 4) ))
- name: api_z1
  instances: 3
  networks:
  - name: net1
    static_ips: (( static_ips(1, 3, 5) ))
```

Next, we'll define our `properties.yml`:

```yml
# examples/static-ips/properties.yml
properties:
  staticIP_servers: (( grab jobs.staticIP_z1.networks.net1.static_ips ))
  api_servers: (( grab jobs.api_z1.networks.net1.static_ips ))
```

And lastly, define our `networks.yml`:

```yml
# examples/static-ips/networks.yml
networks:
- name: net1
  subnets:
  - cloud_properties: random
    static:
    - 192.168.0.2 - 192.168.0.10
```

Merge it all together, and see what we get:

```yml
$ spruce merge jobs.yml properties.yml networks.yml
jobs:
- instances: 3
  name: staticIP_z1
  networks:
  - name: net1
    static_ips:
    - 192.168.0.2
    - 192.168.0.4
    - 192.168.0.6
- instances: 3
  name: api_z1
  networks:
  - name: net1
    static_ips:
    - 192.168.0.3
    - 192.168.0.5
    - 192.168.0.7
networks:
- name: net1
  subnets:
  - cloud_properties: random
    static:
    - 192.168.0.2 - 192.168.0.10
properties:
  api_servers:
  - 192.168.0.3
  - 192.168.0.5
  - 192.168.0.7
  staticIP_servers:
  - 192.168.0.2
  - 192.168.0.4
  - 192.168.0.6
```



<a name="ex-availability-zones"></a>
### Static IPs with Availability Zones

Lets define our `jobs.yml`:

```yml
# examples/availability-zones/jobs.yml
instance_groups:
- name: staticIP
  instances: 3
  azs: [z1,z2]
  networks:
  - name: net1
    static_ips: (( static_ips(0, "z2:2", "z1:3") ))
- name: api
  instances: 3
  azs: [z1]
  networks:
  - name: net1
    static_ips: (( static_ips(1, "z1:4", 5) ))
- name: web
  instances: 3
  networks:
  - name: net1
    static_ips: (( static_ips(9, 10, 11) ))
```

Next, we'll define our `properties.yml`:

```yml
# examples/availability-zones/properties.yml
properties:
  staticIP_servers: (( grab instance_groups.staticIP.networks.net1.static_ips ))
  api_servers: (( grab instance_groups.api.networks.net1.static_ips ))
  web_servers: (( grab instance_groups.web.networks.net1.static_ips ))
```

And lastly, define our `networks.yml`:

```yml
# examples/availability-zones/networks.yml
networks:
- name: net1
  subnets:
  - cloud_properties: random
    az: z1
    static:
    - 192.168.0.1 - 192.168.0.10
  - cloud_properties: random
    az: z2
    static:
    - 192.168.2.1 - 192.168.2.10
```

Merge it all together, and see what we get:

```yml
$ spruce merge jobs.yml properties.yml networks.yml
instance_groups:
- name: staticIP
  instances: 3
  azs: [z1,z2]
  networks:
  - name: net1
    static_ips:
    - 192.168.0.1
    - 192.168.2.3
    - 192.168.0.4
- name: api
  instances: 3
  azs: [z1]
  networks:
  - name: net1
    static_ips:
    - 192.168.0.2
    - 192.168.0.5
    - 192.168.0.6
- name: web
  instances: 3
  networks:
  - name: net1
    static_ips:
    - 192.168.0.10
    - 192.168.2.1
    - 192.168.2.2
networks:
- name: net1
  subnets:
  - cloud_properties: random
    az: z1
    static:
    - 192.168.0.1 - 192.168.0.10
  - cloud_properties: random
    az: z2
    static:
    - 192.168.2.1 - 192.168.2.10
properties:
  api_servers:
  - 192.168.0.2
  - 192.168.0.5
  - 192.168.0.6
  staticIP_servers:
  - 192.168.0.1
  - 192.168.2.3
  - 192.168.0.4
  web_servers:
  - 192.168.0.10
  - 192.168.2.1
  - 192.168.2.2
```



<a name="ex-inject"></a>
### Injecting Subtrees

One of the great things about YAML is the oft-overlooked `<<`
inject operator, which lets you start with a copy of another part
of the YAML tree and override keys, like this:

```yml
# examples/inject/all-in-one.yml
meta:
  template: &template
    color: blue
    size: small

green:
  <<: *template
  color: green
```

Here, `$.green.size` will be `small`, but `$.green.color` stays as
`green`:

```yml
$ spruce merge --prune meta all-in-one.yml
green:
  color: green
  size: small
```

That works great if you are in the same file, but what if you want
to inject data from a different file into your current map and
then override some things?

That's where `(( inject ... ))` really shines.

```yml
# examples/inject/templates.yml
meta:
  template:
    color: blue
    size: small
```

```yml
# examples/inject/green.yml
green:
  woot: (( inject meta.template ))
  color: green
```

```yml
$ spruce merge --prune meta templates.yml green.yml
green:
  color: green
  size: small
```

**Note:** The key used for the `(( inject ... ))` call (in this
case, `woot`) is removed from the final tree as part of the
injection operator.


<a name="no-eval"></a>
### Option --no-eval in merge 

This option can be used to merge several stubs without evaluating the spruce logic inside such as `(( grab ))` or `(( concat ))`. For instance, if you have the following two files: 

```
$ cat first.yml
jobs: 
- name: cell 
  templates:
  - name: garden

$ cat second.yml 
jobs:
- name: cc_bridge
- name: router
- name: cell 
  templates:
  - (( prepend ))
  - name: metron_agent
properties: 
  diego: (( grab meta.diego_enabled ))
  ```

You could merge them simply running: 

```
$ spruce merge --skip-eval first.yml second.yml

jobs:
- name: cell 
  templates:
  - name: metron_agent
  - name: garden
- name: cc_bridge
- name: router

properties: 
  diego: (( grab meta.diego_enabled ))

```

<a name="file"></a>
### File

Sometimes you need to include large blocks of text in your YAML, such as the
body of a configuration file, or a script block.  However, the indentation can
cause issues when that block needs to be edited later, and there's no easy way
to use tools to validate the block.

Using the `(( file ... ))` operator allows you to keep the block in its
natural state to allow for easy viewing, editing and processing, but then add
it to YAML file as needed.  It supports specifying the file either by a string
literal or a reference.

The relative path to the file is based on where spruce is run from.
Alternatively, you can set the `SPRUCE_FILE_BASE_PATH` environment variable to
the desired root that your YAML file uses as the reference to the relative
file paths specified.  You can also specify an absolute path in the YAML

```yml
$ SPRUCE_FILE_BASE_PATH=$HOME/myproj/configs

--- # Source file
server:
  nginx:
    name: nginx.conf
    config_file: (( file server.nginx.name ))
  haproxy:
    name: haproxy.cfg
    config_file: (( file "/haproxy/haproxy.cfg" ))
```

The `server.nginx.config_file` will contain the contents of
`$HOME/myproj/configs/nginx.conf`, while the `server.haproxy.config_file` will
contain the contents of `/haproxy/haproxy.cfg`


<a name="params"></a>
### Parameters

Sometimes, you may want to start with a good starting-point
template, but require other YAML files to provide certain values.
Parameters to the rescue!

```yml
# examples/params/global.yml
disks:
  small: 4096
  medium: 8192
  large:  102400
  networks: (( param "please define the networks" ))
  os:
    - ubuntu
    - centos
    - fedora
```

And then combine that with these local definitions:

```yml
# examples/params/local.yml
disks:
  medium: 16384
  networks:
    - name: public
      range: 10.40.0.0/24
    - name: inside
      range: 10.60.0.0/16
```

This works, but if `local.yml` forgot to specify the top-level
*networks* key, or an error should be emitted.

## Author

Written By Geoff Franks and James Hunt, inspired by [spiff](https://github.com/cloudfoundry-incubator/spiff)

Thanks to Long Nguyen for breaking it repeatedly in the interest
of improvement and quality assurance.

## License

Licensed under [the MIT License](https://github.com/geofffranks/spruce/raw/master/LICENSE)
