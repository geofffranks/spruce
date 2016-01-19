package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/geofffranks/simpleyaml" // FIXME: switch back to smallfish/simpleyaml after https://github.com/smallfish/simpleyaml/pull/1 is merged
	"github.com/voxelbrain/goptions"
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

// VERSION holds the Current version of spruce
var VERSION = "1.0.1" // SED MARKER FOR AUTO VERSION BUMPING
// BUILD holds CURRENT BUILD OF SPRUCE
var BUILD = "master" // updated by build.sh
// DIRTY holds Whether any uncommitted changes were found in the working copy
var DIRTY = "" // updated by build.sh

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
var trace bool
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

// TRACE - Prints out a trace message
func TRACE(format string, args ...interface{}) {
	if trace {
		content := fmt.Sprintf(format, args...)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			lines[i] = "-----> " + line
		}
		content = strings.Join(lines, "\n")
		printfStdErr("%s\n", content)
	}
}

func envFlag(varname string) bool {
	val := os.Getenv(varname)
	return val != "" && strings.ToLower(val) != "false" && val != "0"
}

func main() {
	var options struct {
		Debug     bool `goptions:"-D, --debug, description='Enable debugging'"`
		Trace     bool `goptions:"-T, --trace, description='Enable trace mode debugging (very verbose)'"`
		Version   bool `goptions:"-v, --version, description='Display version information'"`
		Concourse bool `goptions:"--concourse, description='Pre/Post-process YAML for Concourse CI (handles {{ }} quoting)'"`
		Action    goptions.Verbs
		Merge     struct {
			Prune []string           `goptions:"--prune, description='Specify keys to prune from final output (may be specified more than once'"`
			Files goptions.Remainder `goptions:"description='Merges file2.yml through fileN.yml on top of file1.yml'"`
		} `goptions:"merge"`
	}
	getopts(&options)

	if envFlag("DEBUG") {
		debug = true
	}
	if options.Debug {
		debug = options.Debug
	}

	if envFlag("TRACE") {
		trace = true
	}
	if options.Trace {
		trace = options.Trace
	}
	if trace {
		debug = true
	}

	handleConcourseQuoting = options.Concourse

	if options.Version {
		printfStdErr("%s - Version %s (%s%s)\n", os.Args[0], VERSION, BUILD, DIRTY)
		exit(0)
		return
	}

	switch {
	case options.Action == "merge":
		if len(options.Merge.Files) >= 1 {
			root := make(map[interface{}]interface{})

			err := mergeAllDocs(root, options.Merge.Files)
			if err != nil {
				printfStdErr("%s\n", err.Error())
				exit(2)
				return
			}

			ev := &Evaluator{Tree: root}
			err = ev.Run(options.Merge.Prune)
			if err != nil {
				printfStdErr("%s\n", err.Error())
				exit(2)
				return
			}

			TRACE("Converting the following data back to YML:")
			TRACE("%#v", ev.Tree)
			merged, err := yaml.Marshal(ev.Tree)
			if err != nil {
				printfStdErr("Unable to convert merged result back to YAML: %s\nData:\n%#v", err.Error(), ev.Tree)
				exit(2)
				return

			}

			var output string
			if handleConcourseQuoting {
				output = dequoteConcourse(merged)
			} else {
				output = string(merged)
			}
			printfStdOut("%s\n", output)

		} else {
			usage()
			return
		}

	default:
		usage()
		return
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
		data, err := ioutil.ReadFile(path)
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
		TRACE("Current data after processing '%s':\n%s", path, tmpYaml)
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
