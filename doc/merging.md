## Merging Rules

Merging in `spruce` is designed to be pretty intuitive. Files to merge are listed
in-order on the command line. The first file serves as the base to the file structure,
and subsequent files are merged on top, adding when keys are new, replacing when keys
exist.

As `spruce` has grown, additional support for complicated operations has been added,
and multiple phases have been introduced into `spruce` to handle the various tasks.

## Order of Operations

1. Build the Root Document

   The first file is loaded into memory as the root document. Each subsequent file
   specified is merged on top of the root document overwriting, appending, and
   deleting as specified.

   Any [array operators](#what-about-arrays) are evaluated as
   each new document is merged on top of the root document. This allows greater control
   over how data in arrays are merged together (append, prepend, insert, merge, replace).

   If any `(( prune ))` operators are defined, the object they apply to is added to
   a list of data to prune, but the data is unmodified. No other operators are
    evaluated at this time.

2. **Merge Phase**

   In this phase, `(( inject ))` operators are evaluated, to flesh out the root document
   as it had been intended to be operated on in the later phases.

   Since `(( inject ))` happens after the array operators have been evaluated, it is not possible
   to use the array operators when attempting to override an array provided via `(( inject ))`.
   Currently the only supported behavior is that data will be appended to arrays.

3. **Param Phase**

   The root document is scanned for `(( param ))` operators. If any exist at this point,
   it means a property had been defined in a way that required a later file to override it,
   but that did not happen. If any params were found, errors are printed out to the user,
   indicating the missing parameters, and `spruce` exits with the failure.

4. **Eval Phase**

   Unless the `--skip-eval` flag is specified to `spruce`, the root document is scanned
   for [operators][operators], to generate a dependency graph and determine the order in
   which operators will be evaluated. Each operator is evaluated on the root document,
   modifying it in some way. All operators remaining in the document are evaluated at this stage.

5. **Pruning**

   Any parts of the root document marked for pruning via `(( prune ))` operators, or the
   `--prune` flag are deleted.

6. **Cherry Picking**

   If the `--cherry-pick` flag was specified, the relevant datastructures are pulled from
   the root document. The cherry-picked data then replaces the root document to display only
   the requested data.

7. **Output**

   Any errors occurring in the Eval Phase or while Pruning/Cherry Picking  are displayed to
   the user, and `spruce` exits with the failure. If no errors are encountered, `spruce`
   formats the root document as YAML, and prints the output to the user.

## What about arrays?

Merging arrays together is slightly more complicated than merging arbitrary-key-values,
because order matters. To aid in this, `spruce` has specific **array operators** that
are used to tell it how to perform array merges:

- `(( append ))` - Adds the data to the end of the corresponding array in the root document.
- `(( prepend ))` - Inserts the data at the beginning of the corresponding array in the root document.
- `(( insert ))` - Inserts the data before or after a specified index, or object.
- `(( merge ))` - Merges the data on top of an existing array based on a common key. This 
  requires each element to be an object, all with the common key used for merging.
- `(( inline ))` - Merges the data ontop of an existing array, based on the indices of the
  array.
- `(( replace ))` - Removes the existing array, and replaces it with the new one.
- `(( delete ))` - Deletes data at a specific index, or objects identified by the value
  of a specified key.

If no array merge operators are defined, arrays are merged according to the following logic:

1. All elements of the array in the root document, and new document are scanned to see
   if they are objects, and to ensure that each element contains the `name` key. If so,
   `(( merge on name ))` is implied, and elements containing the same name will be merged
   together. Any new elements are appended to the end of the array.
2. If `(( merge on name ))` cannot be done because an element does not contain the `name`
   key, or because one of the elements is not an object, an `(( inline ))` merge is performed.
   However, if the `--fallback-append` flag is specified, a `(( append ))` merge is performed
   instead of the `(( inline ))`.

The array operators are further defined with examples in the [array merging documentation][array-merge].

[array-merge]: https://github.com/geofffranks/spruce/blob/master/doc/array-merging.md
[operators]:   https://github.com/geofffranks/spruce/blob/master/doc/operators.md
