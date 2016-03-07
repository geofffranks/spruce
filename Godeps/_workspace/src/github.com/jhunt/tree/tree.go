package tree

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Node struct {
	Name string
	Sub  []Node
}

func New(name string, sub ...Node) Node {
	return Node{
		Name: name,
		Sub:  sub,
	}
}

func (n Node) render(out io.Writer, prefix string, tail bool) {
	interim := "├── "
	if tail {
		interim = "└── "
	}
	nkids := len(n.Sub)

	ss := strings.Split(strings.Trim(n.Name, "\n"), "\n")
	for _, s := range ss {
		fmt.Fprintf(out, "%s%s%s\n", prefix, interim, s)
		interim = "│   "
		if tail {
			interim = "    "
		}
	}

	for i, c := range n.Sub {
		c.render(out, prefix+interim, i == nkids-1)
	}
}

func (n *Node) Append(child Node) {
	n.Sub = append(n.Sub, child)
}

func (n Node) Draw() string {
	var out bytes.Buffer
	fmt.Fprintf(&out, ".\n")
	n.render(&out, "", true)
	return out.String()
}

func (n Node) flatten(prefix, sep string) []string {
	ss := make([]string, 0)
	if len(n.Sub) == 0 {
		return append(ss, fmt.Sprintf("%s%s", prefix, n.Name))
	}
	for _, k := range n.Sub {
		for _, s := range k.flatten(prefix+n.Name+sep, sep) {
			ss = append(ss, s)
		}
	}
	return ss
}

func (n Node) Paths(sep string) []string {
	return n.flatten("", sep)
}
