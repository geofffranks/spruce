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
