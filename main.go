package main

import (
	"fmt"
	"os"

	"github.com/geofffranks/simpleyaml" // FIXME: switch back to smallfish/simpleyaml after https://github.com/smallfish/simpleyaml/pull/1 is merged
	"github.com/voxelbrain/goptions"
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

// Current version of spruce
var VERSION = "0.8.0" // SED MARKER FOR AUTO VERSION BUMPING

var printfStdOut = func(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}

var printfStdErr = func(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

var getopts = func(o interface{}) {
	err := goptions.Parse(o)
	if err != nil {
		usage()
	}
}

var exit = func(code int) {
	os.Exit(code)
}

var usage = func() {
	goptions.PrintHelp()
	exit(1)
}

var debug bool
var handleConcourseQuoting bool

// DEBUG - Prints out a debug message
func DEBUG(format string, args ...interface{}) {
	if debug {
		content := fmt.Sprintf(format, args...)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			lines[i] = "DEBUG> " + line
		}
		content = strings.Join(lines, "\n")
		printfStdErr("%s\n", content)
	}
}

func main() {
	var options struct {
		Debug     bool `goptions:"-D, --debug, description='Enable debugging'"`
		Version   bool `goptions:"-v, --version, description='Display version information'"`
		Concourse bool `goptions:"--concourse, description='Pre/Post-process YAML for Concourse CI (handles {{ }} quoting)'"`
		Action    goptions.Verbs
		Merge     struct {
			Prune []string           `goptions:"--prune, description='Specify keys to prune from final output (may be specified more than once'"`
			Files goptions.Remainder `goptions:"description='Merges file2.yml through fileN.yml on top of file1.yml'"`
		} `goptions:"merge"`
	}
	getopts(&options)

	if os.Getenv("DEBUG") != "" && strings.ToLower(os.Getenv("DEBUG")) != "false" && os.Getenv("DEBUG") != "0" {
		debug = true
	}
	if options.Debug {
		debug = options.Debug
	}

	handleConcourseQuoting = options.Concourse

	if options.Version {
		printfStdErr("%s - Version %s\n", os.Args[0], VERSION)
		exit(0)
	} else {
		switch {
		case options.Action == "merge":
			if len(options.Merge.Files) >= 1 {
				root := make(map[interface{}]interface{})

				err := mergeAllDocs(root, options.Merge.Files)
				if err != nil {
					printfStdErr("%s\n", err.Error())
					exit(2)
				}

				m := &Merger{}
				m.Visit(root, &StaticIPGenerator{root: root})
				m.Visit(root, &DeReferencer{root: root})
				m.Visit(root, &Concatenator{root: root})

				for _, toPrune := range options.Merge.Prune {
					DEBUG("%s: Pruning", toPrune)
					keys := strings.Split(toPrune, ".")
					if len(keys) > 1 {
						parent := strings.Join(keys[0:(len(keys)-1)], ".")
						target := keys[(len(keys) - 1)]
						node, _ := resolveNode(parent, root)
						if _, ok := node.(map[interface{}]interface{}); ok {
							DEBUG("   PRUNED %q from %q", target, parent)
							delete(node.(map[interface{}]interface{}), target)
						}
					} else {
						DEBUG("   PRUNED %q from \"$\"", keys[0])
						delete(root, keys[0])
					}
				}

				m.Visit(root, &ParamChecker{})

				if err := m.Error(); err != nil {
					printfStdErr("%s\n", err.Error())
					exit(2)

				} else {
					DEBUG("Converting the following data back to YML:")
					DEBUG("%#v", root)
					merged, err := yaml.Marshal(root)
					if err != nil {
						printfStdErr("Unable to convert merged result back to YAML: %s\nData:\n%#v", err.Error(), root)
						exit(2)
					} else {
						var output string
						if handleConcourseQuoting {
							output = dequoteConcourse(merged)
						} else {
							output = string(merged)
						}
						printfStdOut("%s\n", output)
					}
				}
			} else {
				usage()
			}
		default:
			usage()
		}
	}
}

func parseYAML(data []byte) (map[interface{}]interface{}, error) {
	y, err := simpleyaml.NewYaml(data)
	if err != nil {
		return nil, err
	}

	doc, err := y.Map()
	if err != nil {
		return nil, fmt.Errorf("Root of YAML document is not a hash/map: %s\n", err.Error())
	}

	return doc, nil
}

func mergeAllDocs(root map[interface{}]interface{}, paths []string) error {
	m := &Merger{}

	for _, path := range paths {
		DEBUG("Processing file '%s'", path)
		data, err := readFile(path)
		if err != nil {
			return fmt.Errorf("Error reading file %s: %s\n", path, err.Error())
		}

		if handleConcourseQuoting {
			data = quoteConcourse(data)
		}

		doc, err := parseYAML(data)
		if err != nil {
			return fmt.Errorf("%s: %s\n", path, err.Error())
		}

		m.Merge(root, doc)

		tmpYaml, _ := yaml.Marshal(root) // we don't care about errors for debugging
		DEBUG("Current data after processing '%s':\n%s", path, tmpYaml)
	}

	return m.Error()
}

var concourseRegex = `\{\{([-\w\p{L}]+)\}\}`

func quoteConcourse(input []byte) []byte {
	re := regexp.MustCompile("(" + concourseRegex + ")")
	return re.ReplaceAll(input, []byte("\"$1\""))
}

func dequoteConcourse(input []byte) string {
	re := regexp.MustCompile("['\"](" + concourseRegex + ")[\"']")
	return re.ReplaceAllString(string(input), "$1")
}
