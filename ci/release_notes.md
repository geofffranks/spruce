# New Features
1. Added support for requiring values to be specified in a later template. This
   was something that you could do implicitly with `spiff`, via (( merge )), and getting
   failure if you didn't define the value. In `spruce`, this is now something that you can
   explicitly require from downstream templates:

   ```(( param "Your custom error message here" ))```

2. Added support for concatenating values together, either strings, references, or both:

   ``` (( concat properties.myjob.protocol properties.myjob.host ":" properties.myjob.port ```

   Concatenation is done after dereferencing, in case any of the properties reference something like
   a static_ip from another node.

3. Made merging arrays the default behavior (previously, they replaced by default). Since
   everything else merged by default, and most cases want merging this just made sense. When
   merging arrays, `spruce` will try to do a key-based merge, on the `name` key, and failing that,
   does an index-based merge.

# Bug Fixes

1. Fixed issue resulting in a panic if specifying `static_ips(0)` - this should have been a 0-based
   index lookup for greater compatibility with spiff templates.
2. Fixed an issue where you could not resolve a static IP defined with `static_ips()`, when
   targeting specific elements in the array - `(( jobs.myjob.networks.mynet.static_ips ))` worked,
   but `(( jobs.myjob.networks.mynet.static_ips.[0] ))` did not. It now does. Yay!
3. Fixed an issue where a panic would occur during postprocessing of keys that had null (`~`) values). Oops!

# Acknowledgements

Thanks to [James Hunt](https://github.com/filefrog) for the hard work on param support, array-merge-by-default,
value concatenation, and the nil-reference panic bugfix!

Thanks to [Long Nguyen](https://github.com/longnguyen11288) for all the bug reports + field testing!
