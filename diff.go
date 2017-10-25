package spruce

import (
	"regexp"
	"sort"
	"strings"

	"github.com/geofffranks/yaml"
	fmt "github.com/starkandwayne/goutils/ansi"
)

func pad1(pad, s string) string {
	return strings.TrimSpace(indent(pad, s))
}

func indent(pad, s string) string {
	re := regexp.MustCompile(`(?m)^`)
	return re.ReplaceAllString(s, pad)
}

func yamlstring(f, pad string, x Diffable) string {
	s, _ := yaml.Marshal(x.Value())
	return fmt.Sprintf(f, indent(pad, strings.TrimSuffix(string(s), "\n")))
}

func yamlmarshal(x interface{}) string {
	s, _ := yaml.Marshal(x)
	return fmt.Sprintf("%s", s)
}

func sortkeys(m map[string]Diffable) []string {
	kk := make([]string, 0)
	for k := range m {
		kk = append(kk, k)
	}
	sort.Strings(kk)
	return kk
}

type Type int

const (
	Scalar Type = iota
	Map
	SimpleList
	KeyedList
)

func keyed(l []interface{}) string {
	for _, v := range l {
		if typeof(v) != Map {
			return ""
		}
	}

KEYSEARCH:
	for _, k := range []string{"name", "id", "key"} {
		for _, v := range l {
			o := v.(map[interface{}]interface{})
			if _, ok := o[k]; !ok {
				continue KEYSEARCH
			}
		}

		return k
	}

	return ""
}

func mapify(l []interface{}, key string) map[interface{}]interface{} {
	m := make(map[interface{}]interface{})

	for _, v := range l {
		if typeof(v) != Map {
			return nil
		}
		o := v.(map[interface{}]interface{})
		k, ok := o[key]
		if !ok {
			return nil
		}
		m[k] = v
	}

	return m
}

func (t Type) String() string {
	switch t {
	case Scalar:
		return "scalar"
	case Map:
		return "map"
	case SimpleList:
		return "simple list"
	case KeyedList:
		return "keyed list"
	default:
		return "unknown type"
	}
}

func typeof(x interface{}) Type {
	switch x.(type) {
	case map[interface{}]interface{}:
		return Map
	case []interface{}:
		l := x.([]interface{})
		if keyed(l) != "" {
			return KeyedList
		}
		return SimpleList

	default:
		return Scalar
	}
}

type Diffable interface {
	Changed() bool
	String(key string) string
	Value() interface{}
}

type DiffNone struct {
	Orig interface{}
}

func (d DiffNone) Changed() bool {
	return false
}

func (d DiffNone) String(key string) string {
	return ""
}

func (d DiffNone) Value() interface{} {
	return d.Orig
}

type DiffType struct {
	Old interface{}
	New interface{}
}

func (d DiffType) Changed() bool {
	return typeof(d.Old) != typeof(d.New)
}

func (d DiffType) String(key string) string {
	return fmt.Sprintf("  @C{%s} changed type\n    from @R{%s}\n      to @G{%s}\n\n",
		key, typeof(d.Old), typeof(d.New))
}

func (d DiffType) Value() interface{} {
	return nil
}

type DiffScalar struct {
	Old string
	New string
}

func (d DiffScalar) Changed() bool {
	return d.Old != d.New
}

func (d DiffScalar) String(key string) string {
	return fmt.Sprintf("  @C{%s} changed value\n    from @R{%s}\n      to @G{%s}\n\n",
		key, pad1("         ", d.Old), pad1("         ", d.New))
}

func (d DiffScalar) Value() interface{} {
	return nil
}

type DiffMap struct {
	Removed map[string]Diffable
	Added   map[string]Diffable
	Common  map[string]Diffable
}

func (d DiffMap) Changed() bool {
	if len(d.Removed)+len(d.Added) > 0 {
		return true
	}

	for _, x := range d.Common {
		if x.Changed() {
			return true
		}
	}
	return false
}

func (d DiffMap) String(key string) string {
	s := ""

	for _, k := range sortkeys(d.Added) {
		v := d.Added[k]
		s += fmt.Sprintf("  @C{%s.%s} added\n", key, k)
		s += yamlstring("@G{%s}\n\n", "    ", v)
	}
	for _, k := range sortkeys(d.Removed) {
		v := d.Removed[k]
		s += fmt.Sprintf("  @C{%s.%s} removed\n", key, k)
		s += yamlstring("@R{%s}\n\n", "    ", v)
	}

	for _, k := range sortkeys(d.Common) {
		v := d.Common[k]
		if v.Changed() {
			s += v.String(fmt.Sprintf("%s.%v", key, k))
		}
	}
	return s
}

func (d DiffMap) Value() interface{} {
	return nil
}

type DiffList struct {
	Removed map[string]Diffable
	Added   map[string]Diffable
	Common  map[string]Diffable
}

func (d DiffList) Changed() bool {
	if len(d.Removed)+len(d.Added) > 0 {
		return true
	}
	for _, x := range d.Common {
		if x.Changed() {
			return true
		}
	}
	return false
}

func (d DiffList) String(key string) string {
	s := ""

	for _, k := range sortkeys(d.Added) {
		v := d.Added[k]
		s += fmt.Sprintf("  @C{%s[%s]} added\n", key, k)
		s += yamlstring("@G{%s}\n\n", "    ", v)
	}
	for _, k := range sortkeys(d.Removed) {
		v := d.Removed[k]
		s += fmt.Sprintf("  @C{%s[%s]} removed\n", key, k)
		s += yamlstring("@R{%s}\n\n", "    ", v)
	}
	for _, k := range sortkeys(d.Common) {
		v := d.Common[k]
		if v.Changed() {
			s += v.String(fmt.Sprintf("%s[%s]", key, k))
		}
	}
	return s
}

func (d DiffList) Value() interface{} {
	return nil
}

func Diff(a, b interface{}) (Diffable, error) {
	if typeof(a) != typeof(b) {
		return DiffType{
			Old: a,
			New: b,
		}, nil
	}

	switch typeof(a) {
	case Scalar:
		return DiffScalar{
			Old: yamlmarshal(a),
			New: yamlmarshal(b),
		}, nil

	case Map:
		ma := a.(map[interface{}]interface{})
		mb := b.(map[interface{}]interface{})
		x := DiffMap{
			Removed: make(map[string]Diffable),
			Added:   make(map[string]Diffable),
			Common:  make(map[string]Diffable),
		}
		for k, v1 := range ma {
			if v2, ok := mb[k]; ok {
				d, err := Diff(v1, v2)
				if err != nil {
					return x, err
				}
				x.Common[fmt.Sprintf("%v", k)] = d
				continue
			}

			x.Removed[fmt.Sprintf("%v", k)] = DiffNone{v1}
			continue
		}

		for k, v2 := range mb {
			if _, ok := ma[k]; ok {
				continue
			}

			x.Added[fmt.Sprintf("%v", k)] = DiffNone{v2}
			continue
		}
		return x, nil

	case SimpleList:
		la := a.([]interface{})
		lb := b.([]interface{})
		x := DiffList{
			Removed: make(map[string]Diffable),
			Added:   make(map[string]Diffable),
			Common:  make(map[string]Diffable),
		}
		for i, v1 := range la {
			if i < len(lb) {
				v2 := lb[i]
				d, err := Diff(v1, v2)
				if err != nil {
					return x, err
				}
				x.Common[fmt.Sprintf("%d", i)] = d
				continue
			}

			x.Removed[fmt.Sprintf("%d", i)] = DiffNone{v1}
			continue
		}

		for i, v2 := range lb {
			if i < len(la) {
				continue
			}

			x.Added[fmt.Sprintf("%d", i)] = DiffNone{v2}
			continue
		}
		return x, nil

	case KeyedList:
		la := a.([]interface{})
		lb := b.([]interface{})
		key := keyed(la)

		ma := mapify(la, key)
		mb := mapify(lb, key)

		x := DiffList{
			Removed: make(map[string]Diffable),
			Added:   make(map[string]Diffable),
			Common:  make(map[string]Diffable),
		}

		for k, v1 := range ma {
			if v2, ok := mb[k]; ok {
				d, err := Diff(v1, v2)
				if err != nil {
					return x, err
				}
				x.Common[fmt.Sprintf("%v", k)] = d
				continue
			}

			x.Removed[fmt.Sprintf("%v", k)] = DiffNone{v1}
			continue
		}

		for k, v2 := range mb {
			if _, ok := ma[k]; ok {
				continue
			}

			x.Added[fmt.Sprintf("%v", k)] = DiffNone{v2}
			continue
		}
		return x, nil

	default:
		return DiffScalar{}, fmt.Errorf("not implemented yet!")
	}
}
