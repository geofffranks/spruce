package tree_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/starkandwayne/goutils/tree"
)

var trim = regexp.MustCompile("\t+")

func drawsOk(t *testing.T, msg string, n tree.Node, s string) {
	got := strings.Trim(n.Draw(), "\n")
	want := strings.Trim(trim.ReplaceAllString(s, ""), "\n")
	if got != want {
		t.Errorf("%s failed\nexpected:\n[%s]\ngot:\n[%s]\n", msg, want, got)
	}
}

func pathsOk(t *testing.T, msg string, n tree.Node, want ...string) {
	got := n.Paths("/")
	if len(got) != len(want) {
		got_ := "    - " + strings.Join(got, "\n    - ") + "\n"
		want_ := "    - " + strings.Join(want, "\n    - ") + "\n"
		t.Errorf("%s failed\nexpected %d paths:\n%s\ngot %d paths:\n%s\n", msg, len(want), want_, len(got), got_)
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s failed\npaths[%d] was incorrect\nexpected: [%s]\n     got: [%s]\n", msg, i, want[i], got[i])
		}
	}
}

func pathSegmentsOk(t *testing.T, msg string, n tree.Node, want [][]string) {
	got := n.PathSegments()
	if len(got) != len(want) {
		t.Errorf("%s failed; wanted %d distinct , got %d:\n\nwanted:\n%v\n\ngot:\n%v", msg, len(want), len(got), want, got)
	}
	for i := range got {
		if len(got[i]) != len(want[i]) {
			t.Errorf("%s failed; wanted path #%d to have %d segments, got %d:\n\nwanted: %v\n\ngot: %v", msg, i, len(want[i]), len(got[i]), want[i], got[i])
		}
		for j := range got[i] {
			if got[i][j] != want[i][j] {
				t.Errorf("%s failed; segment %d of path #%d expected '%s', got '%s'", msg, j, i, want[i][j], got[i][j])
			}
		}
	}
}

func TestDrawing(t *testing.T) {
	drawsOk(t, "{a}",
		tree.New("a"), `
		.
		└── a`)

	drawsOk(t, "{a -> b}",
		tree.New("a", tree.New("b")), `
		.
		└── a
		    └── b`)

	drawsOk(t, "{a -> [b c]}",
		tree.New("a", tree.New("b"), tree.New("c")), `
		.
		└── a
		    ├── b
		    └── c`)

	drawsOk(t, "{a -> b -> c}",
		tree.New("a", tree.New("b", tree.New("c"))), `
		.
		└── a
		    └── b
		        └── c`)

	drawsOk(t, "{a -> [{b -> c} d]}",
		tree.New("a", tree.New("b", tree.New("c")), tree.New("d")), `
		.
		└── a
		    ├── b
		    │   └── c
		    └── d`)

	drawsOk(t, "{a -> [{b -> c -> e} d]}",
		tree.New("a", tree.New("b", tree.New("c", tree.New("e"))), tree.New("d")), `
		.
		└── a
		    ├── b
		    │   └── c
		    │       └── e
		    └── d`)

	drawsOk(t, "multiline node strings",
		tree.New("Alpha\n(first)\n",
			tree.New("Beta\n(second)\n",
				tree.New("Gamma\n(third)\n"),
			),
			tree.New("Delta\n(fourth)\n"),
		), `
		.
		└── Alpha
		    (first)
		    ├── Beta
		    │   (second)
		    │   └── Gamma
		    │       (third)
		    └── Delta
		        (fourth)`)
}

func TestPaths(t *testing.T) {
	pathsOk(t, "{a}",
		tree.New("a"),
		"a")

	pathsOk(t, "{a -> b}",
		tree.New("a", tree.New("b")),
		"a/b")

	pathsOk(t, "{a -> [b c]}",
		tree.New("a", tree.New("b"), tree.New("c")),
		"a/b",
		"a/c")

	pathsOk(t, "{a -> [{b -> c} d]",
		tree.New("a", tree.New("b", tree.New("c")), tree.New("d")),
		"a/b/c",
		"a/d")
}

func TestPathSegments(t *testing.T) {
	pathSegmentsOk(t, "complicated nested structure",
		tree.New(".",
			tree.New("a",
				tree.New("a1"),
				tree.New("a2")),
			tree.New("b",
				tree.New("b1",
					tree.New("b1i"),
					tree.New("b1ii")),
				tree.New("b2")),
			tree.New("c")),
		[][]string{
			[]string{".", "a", "a1"},
			[]string{".", "a", "a2"},
			[]string{".", "b", "b1", "b1i"},
			[]string{".", "b", "b1", "b1ii"},
			[]string{".", "b", "b2"},
			[]string{".", "c"},
		},
	)
}

func TestAppend(t *testing.T) {
	tr := tree.New("a", tree.New("b"))
	drawsOk(t, "{a -> b} before append", tr, `
		.
		└── a
		    └── b`)

	tr.Append(tree.New("c"))
	drawsOk(t, "{a -> [b c]} after append", tr, `
		.
		└── a
		    ├── b
		    └── c`)
}
