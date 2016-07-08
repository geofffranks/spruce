tree - A Library for Printing Tree Structures
=============================================

`tree` makes it tree-vial to print out tree structures,
or walk through them and retrieve data from arbitrary leaves.


## Showing the tree
```go
package main

import (
	"fmt"
	"github.com/jhunt/tree"
)

func main() {
	t := tree.New("a",
		tree.New("b"),
		tree.New("c"),
	)

	fmt.Printf("%s\n", t.Draw())
}
```

Will print out a nicely formatted tree, using ANSI line graphics:

```
.
└── a
    ├── b
    └── c
```


